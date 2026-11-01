package data

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// JSONScriptRepository stores the saved scripts as a JSON array in one file.
type JSONScriptRepository struct {
	store FileStore
	path  string
	now   func() time.Time
}

// NewJSONScriptRepository builds a repository that keeps its list in path, through store, dating additions with now.
func NewJSONScriptRepository(store FileStore, path string, now func() time.Time) *JSONScriptRepository {
	return &JSONScriptRepository{store: store, path: path, now: now}
}

// List returns the saved scripts; a missing file is an empty list, a corrupt one is an error.
func (repository *JSONScriptRepository) List() ([]domain.Script, error) {
	var scripts []domain.Script
	if err := readJSON(repository.store, repository.path, "saved scripts", &scripts); err != nil {
		return nil, err
	}
	return scripts, nil
}

// Add saves the script at path, refusing a path that is already saved.
func (repository *JSONScriptRepository) Add(path string) (domain.Script, error) {
	scripts, err := repository.List()
	if err != nil {
		return domain.Script{}, err
	}
	script := domain.NewScript(path, repository.now())
	if slices.ContainsFunc(scripts, func(saved domain.Script) bool { return saved.ID == script.ID }) {
		return domain.Script{}, fmt.Errorf("%w: %s", ErrDuplicateScript, script.Path)
	}
	return script, repository.save(append(scripts, script))
}

// Remove forgets the script with this ID.
func (repository *JSONScriptRepository) Remove(id string) error {
	scripts, err := repository.List()
	if err != nil {
		return err
	}
	return repository.save(slices.DeleteFunc(scripts, func(saved domain.Script) bool { return saved.ID == id }))
}

func (repository *JSONScriptRepository) save(scripts []domain.Script) error {
	return writeJSON(repository.store, repository.path, "saved scripts", scripts)
}

// OSFileStore is a FileStore on the real disk.
type OSFileStore struct{}

// ReadFile reads the whole file from disk.
func (OSFileStore) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

// WriteFile writes through a temporary file and a rename, so a crash never leaves half a file.
func (OSFileStore) WriteFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, content, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

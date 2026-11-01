package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
	content, err := repository.store.ReadFile(repository.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading saved scripts from %s: %w", repository.path, err)
	}
	var scripts []domain.Script
	if err := json.Unmarshal(content, &scripts); err != nil {
		return nil, fmt.Errorf("saved scripts file %s is corrupt, fix or delete it: %w", repository.path, err)
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
	content, err := json.MarshalIndent(scripts, "", "\t")
	if err != nil {
		return fmt.Errorf("encoding saved scripts: %w", err)
	}
	if err := repository.store.WriteFile(repository.path, content); err != nil {
		return fmt.Errorf("writing saved scripts to %s: %w", repository.path, err)
	}
	return nil
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

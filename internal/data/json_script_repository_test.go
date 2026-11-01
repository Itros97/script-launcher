package data

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"
)

const scriptsFile = "/config/scripts.json"

// memoryStore is a FileStore kept in a map.
type memoryStore map[string][]byte

func (store memoryStore) ReadFile(path string) ([]byte, error) {
	content, ok := store[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return content, nil
}

func (store memoryStore) WriteFile(path string, content []byte) error {
	store[path] = content
	return nil
}

func fixedClock() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }

func newRepository(store memoryStore) *JSONScriptRepository {
	return NewJSONScriptRepository(store, scriptsFile, fixedClock)
}

func TestListIsEmptyWhenNothingWasSaved(t *testing.T) {
	scripts, err := newRepository(memoryStore{}).List()
	if err != nil || len(scripts) != 0 {
		t.Fatalf("got %v, %v", scripts, err)
	}
}

func TestAddedScriptsSurviveANewRepository(t *testing.T) {
	store := memoryStore{}
	if _, err := newRepository(store).Add("/scripts/a.sh"); err != nil {
		t.Fatal(err)
	}
	if _, err := newRepository(store).Add("/scripts/b.py"); err != nil {
		t.Fatal(err)
	}
	scripts, err := newRepository(store).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 2 || scripts[0].Name != "a.sh" || scripts[1].Name != "b.py" || !scripts[0].AddedAt.Equal(fixedClock()) {
		t.Errorf("got %+v", scripts)
	}
}

func TestRefusesToAddTheSamePathTwice(t *testing.T) {
	repository := newRepository(memoryStore{})
	if _, err := repository.Add("/scripts/a.sh"); err != nil {
		t.Fatal(err)
	}
	_, err := repository.Add(filepath.Join("/scripts", ".", "a.sh"))
	if !errors.Is(err, ErrDuplicateScript) {
		t.Fatalf("got %v, want ErrDuplicateScript", err)
	}
}

func TestRemoveForgetsOnlyThatScript(t *testing.T) {
	repository := newRepository(memoryStore{})
	first, _ := repository.Add("/scripts/a.sh")
	_, _ = repository.Add("/scripts/b.sh")
	if err := repository.Remove(first.ID); err != nil {
		t.Fatal(err)
	}
	scripts, _ := repository.List()
	if len(scripts) != 1 || scripts[0].Name != "b.sh" {
		t.Errorf("got %+v", scripts)
	}
}

func TestCorruptFileIsAnErrorAndIsNotOverwritten(t *testing.T) {
	store := memoryStore{scriptsFile: []byte("{not json")}
	repository := newRepository(store)
	if _, err := repository.List(); err == nil {
		t.Fatal("expected an error for a corrupt file")
	}
	if _, err := repository.Add("/scripts/a.sh"); err == nil {
		t.Fatal("expected Add to fail on a corrupt file")
	}
	if string(store[scriptsFile]) != "{not json" {
		t.Errorf("corrupt file was overwritten: %q", store[scriptsFile])
	}
}

func TestOSFileStoreCreatesFolderAndRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "scripts.json")
	store := OSFileStore{}
	if err := store.WriteFile(path, []byte("[]")); err != nil {
		t.Fatal(err)
	}
	content, err := store.ReadFile(path)
	if err != nil || string(content) != "[]" {
		t.Errorf("got %q, %v", content, err)
	}
}

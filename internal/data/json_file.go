package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
)

// readJSON decodes the file at path into target, leaving target alone when the file does not exist.
// A corrupt file is an error naming what it holds, so it is fixed rather than overwritten.
func readJSON(store FileStore, path, what string, target any) error {
	content, err := store.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s from %s: %w", what, path, err)
	}
	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("%s file %s is corrupt, fix or delete it: %w", what, path, err)
	}
	return nil
}

// writeJSON encodes value with tabs and replaces the file at path.
func writeJSON(store FileStore, path, what string, value any) error {
	content, err := json.MarshalIndent(value, "", "\t")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", what, err)
	}
	if err := store.WriteFile(path, content); err != nil {
		return fmt.Errorf("writing %s to %s: %w", what, path, err)
	}
	return nil
}

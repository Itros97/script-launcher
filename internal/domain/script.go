package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"time"
)

// Script is a file the user saved to run again without looking for it.
type Script struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	AddedAt time.Time `json:"addedAt"`
}

// NewScript builds a Script for path, with an ID derived from the path so the same file always gets the same ID.
func NewScript(path string, addedAt time.Time) Script {
	clean := filepath.Clean(path)
	return Script{ID: ScriptID(clean), Name: filepath.Base(clean), Path: clean, AddedAt: addedAt}
}

// ScriptID returns a short stable identifier for a script path.
func ScriptID(path string) string {
	sum := sha1.Sum([]byte(filepath.Clean(path)))
	return hex.EncodeToString(sum[:6])
}

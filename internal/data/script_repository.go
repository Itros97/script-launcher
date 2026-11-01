package data

import (
	"errors"

	"github.com/Itros97/script-launcher/internal/domain"
)

// ErrDuplicateScript is returned when a path is already saved.
var ErrDuplicateScript = errors.New("script already saved")

// ScriptRepository keeps the list of saved scripts between runs of the app.
type ScriptRepository interface {
	// List returns the saved scripts in the order they were added.
	List() ([]domain.Script, error)
	// Add saves the script at path and returns it.
	Add(path string) (domain.Script, error)
	// Remove forgets the script with this ID; removing an unknown ID is not an error.
	Remove(id string) error
}

// FileStore reads and writes whole files; it hides the disk so the repository can be tested in memory.
type FileStore interface {
	// ReadFile returns the file's bytes, or an error wrapping fs.ErrNotExist when it is missing.
	ReadFile(path string) ([]byte, error)
	// WriteFile replaces the file's content, creating its folder if needed.
	WriteFile(path string, content []byte) error
}

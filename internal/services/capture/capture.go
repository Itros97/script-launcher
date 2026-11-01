// Package capture embeds the Python scripts that use the RealSense camera: capture.py saves one
// photo, preview.py streams frames, and camera.py is what both share.
package capture

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Script names, as written by WriteScripts.
const (
	CaptureScript = "capture.py"
	PreviewScript = "preview.py"
)

//go:embed *.py
var scripts embed.FS

// WriteScripts puts every script in dir, where the project's Python can run them, replacing older copies.
func WriteScripts(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	names, err := fs.Glob(scripts, "*.py")
	if err != nil {
		return err
	}
	for _, name := range names {
		content, err := scripts.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	return nil
}

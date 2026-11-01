package services

import (
	"os"
	"os/exec"
)

// macPythons are where the python.org installer and Homebrew put python3. They are tried before
// PATH because an app opened from the Finder gets a short PATH whose python3 is Apple's stand-in,
// which only offers to install the developer tools.
var macPythons = []string{
	"/Library/Frameworks/Python.framework/Versions/Current/bin/python3",
	"/opt/homebrew/bin/python3",
	"/usr/local/bin/python3",
}

// FindPython returns the python3 to create environments with on the system goos: on macOS the
// first installed one of macPythons, elsewhere (or when none is) the one in PATH, or "python3" so
// the error names what is missing.
func FindPython(goos string, exists func(path string) bool, lookPath func(name string) (string, error)) string {
	if goos == "darwin" {
		for _, path := range macPythons {
			if exists(path) {
				return path
			}
		}
	}
	if path, err := lookPath("python3"); err == nil {
		return path
	}
	return "python3"
}

// SystemPython is FindPython for this computer.
func SystemPython(goos string) string {
	return FindPython(goos, isFile, exec.LookPath)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

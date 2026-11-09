package views

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// fakePythonDownload packs a python/bin/python3 that hands over to the real python3.
func fakePythonDownload(t *testing.T) domain.PythonDownload {
	t.Helper()
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "python", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "python", "bin", "python3"), []byte("#!/bin/sh\nexec python3 \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "python.tar.gz")
	if output, err := exec.Command("tar", "-czf", archive, "-C", src, "python").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	content, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	return domain.PythonDownload{Version: "3.12.0", URL: "file://" + archive, SHA256: hex.EncodeToString(sum[:]), SizeMB: 1}
}

func TestPythonButtonInstallsPython(t *testing.T) {
	pythonDir := filepath.Join(t.TempDir(), "python-3.12.0")
	download := fakePythonDownload(t)
	app := openAppWith(t, data.Settings{Profile: domain.Calibre3DProfileID, Calibre3DProject: calibreProject(t)}, func(env *services.PythonEnvService) {
		env.WithDownload(download, pythonDir)
	})
	waitForText(t, app.window, "Sin Python")
	click(t, app.window, "Instalar Python 3.12.0")
	waitForText(t, app.window, "Python 3.12.0 listo")
	waitForText(t, app.window, "✓ Python instalado")
	if _, err := os.Stat(domain.DownloadedPython(pythonDir)); err != nil {
		t.Errorf("Python not installed: %v", err)
	}
}

func TestPythonCardIsHiddenWithoutADownload(t *testing.T) {
	app := openCalibre(t, calibreProject(t))
	waitForText(t, app.window, "Sin entorno")
	if slices.Contains(app.window.Texts(), "Sin Python") {
		t.Error("the Python card should not be shown without a download")
	}
}

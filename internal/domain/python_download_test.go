package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeArchive packs a python/bin/python3 that prints "fake" and returns its download.
func fakeArchive(t *testing.T) PythonDownload {
	t.Helper()
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "python", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "python", "bin", "python3"), []byte("#!/bin/sh\necho fake\n"), 0o755); err != nil {
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
	return PythonDownload{Version: "3.12.0", URL: "file://" + archive, SHA256: hex.EncodeToString(sum[:]), SizeMB: 1}
}

func TestThereIsADownloadForEveryComputerTheLauncherShipsFor(t *testing.T) {
	for _, target := range [][2]string{{"linux", "amd64"}, {"darwin", "arm64"}, {"darwin", "amd64"}} {
		download, ok := PythonDownloadFor(target[0], target[1])
		if !ok || !strings.HasPrefix(download.URL, "https://") || len(download.SHA256) != 64 {
			t.Errorf("%v: got %+v, %v", target, download, ok)
		}
	}
	if _, ok := PythonDownloadFor("windows", "amd64"); ok {
		t.Error("windows should have no download")
	}
}

func TestPythonInstallStepsUnpackTheCheckedArchive(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "python-3.12.0")
	script := SetupScript(PythonInstallSteps(fakeArchive(t), dir))
	if output, err := exec.Command("bash", "-c", script).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	output, err := exec.Command(DownloadedPython(dir)).Output()
	if err != nil || string(output) != "fake\n" {
		t.Errorf("got %q, %v", output, err)
	}
	if leftovers, _ := filepath.Glob(dir + ".*"); len(leftovers) != 0 {
		t.Errorf("left behind %v", leftovers)
	}
}

func TestPythonInstallStepsRefuseADamagedDownload(t *testing.T) {
	download := fakeArchive(t)
	download.SHA256 = strings.Repeat("0", 64)
	dir := filepath.Join(t.TempDir(), "python-3.12.0")
	output, err := exec.Command("bash", "-c", SetupScript(PythonInstallSteps(download, dir))).CombinedOutput()
	if err == nil || !strings.Contains(string(output), "dañada") {
		t.Errorf("got %q, %v", output, err)
	}
	if _, err := os.Stat(DownloadedPython(dir)); err == nil {
		t.Error("a damaged download was installed")
	}
}

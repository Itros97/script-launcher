package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

const fakeModule = "fakecalibre"

// fakeProject creates a project folder with an importable module and no requirements.
func fakeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, fakeModule), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, fakeModule, "__init__.py"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newEnvService(python string) *PythonEnvService {
	return NewPythonEnvService(newService(), python, fakeModule)
}

func checkStatus(t *testing.T, service *PythonEnvService, dir string) domain.EnvStatus {
	t.Helper()
	check, err := service.Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	return check.Status
}

func TestCheckSaysThereIsNoProjectWithoutAFolder(t *testing.T) {
	service := newEnvService("python3")
	for _, dir := range []string{"", filepath.Join(t.TempDir(), "gone")} {
		if status := checkStatus(t, service, dir); status != domain.NoProject {
			t.Errorf("%q: got %d, want NoProject", dir, status)
		}
	}
}

func TestCheckSaysThereIsNoVenvInAFreshProject(t *testing.T) {
	if status := checkStatus(t, newEnvService("python3"), fakeProject(t)); status != domain.NoVenv {
		t.Errorf("got %d, want NoVenv", status)
	}
}

func TestSetupCreatesAVenvThatImportsTheModule(t *testing.T) {
	service := newEnvService("python3")
	dir := fakeProject(t)
	run, err := service.Setup(dir)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := waitFor(t, run); snapshot.Status != Succeeded {
		t.Fatalf("setup failed: %+v", snapshot.Lines)
	}
	if latest, ok := service.LatestSetup(dir); !ok || latest != run {
		t.Error("LatestSetup does not return the setup run")
	}
	if status := checkStatus(t, service, dir); status != domain.Ready {
		t.Errorf("got %d, want Ready", status)
	}
	if err := os.RemoveAll(filepath.Join(dir, fakeModule)); err != nil {
		t.Fatal(err)
	}
	check, err := service.Check(dir)
	if err != nil || check.Status != domain.Incomplete || !strings.Contains(check.Reason, "En esta carpeta no está el programa "+fakeModule) {
		t.Errorf("without the module: got %+v, %v", check, err)
	}
}

func TestSetupFailsClearlyWithoutPython(t *testing.T) {
	_, err := newEnvService("python-that-does-not-exist").Setup(fakeProject(t))
	if !errors.Is(err, ErrNoPython) {
		t.Errorf("got %v, want ErrNoPython", err)
	}
}

// fakePythonDownload packs a python/bin/python3 that hands over to the real python3, so the
// environment it creates works.
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

func TestSetupDownloadsPythonWhenTheComputerHasNone(t *testing.T) {
	pythonDir := filepath.Join(t.TempDir(), "python-3.12.0")
	service := newEnvService("python-that-does-not-exist").WithDownload(fakePythonDownload(t), pythonDir)
	dir := fakeProject(t)
	run, err := service.Setup(dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Succeeded {
		t.Fatalf("setup failed: %+v", snapshot.Lines)
	}
	if !strings.Contains(snapshot.Lines[0].Text, "descargando Python 3.12.0") {
		t.Errorf("first line: %q", snapshot.Lines[0].Text)
	}
	if status := checkStatus(t, service, dir); status != domain.Ready {
		t.Errorf("got %d, want Ready", status)
	}
}

func TestSetupKeepsUsingADownloadedPython(t *testing.T) {
	pythonDir := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(domain.DownloadedPython(pythonDir)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(domain.DownloadedPython(pythonDir), []byte("#!/bin/sh\nexec python3 \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	download := domain.PythonDownload{Version: "3.12.0", URL: "file:///nowhere", SHA256: strings.Repeat("0", 64)}
	service := newEnvService("python3").WithDownload(download, pythonDir)
	run, err := service.Setup(fakeProject(t))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Succeeded || strings.Contains(snapshot.Lines[0].Text, "descargando") {
		t.Errorf("got %+v", snapshot.Lines)
	}
}

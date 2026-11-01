package services

import (
	"errors"
	"os"
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

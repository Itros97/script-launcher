package views

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// readyProject creates a calibre3d project with a real .venv whose pyrealsense2 is the services' fake camera.
func readyProject(t *testing.T) string {
	t.Helper()
	project := calibreProject(t)
	if output, err := exec.Command("python3", "-m", "venv", filepath.Join(project, domain.VenvDir)).CombinedOutput(); err != nil {
		t.Fatalf("creating venv: %v\n%s", err, output)
	}
	sitePackages, err := filepath.Glob(filepath.Join(project, domain.VenvDir, "lib", "python3*", "site-packages"))
	if err != nil || len(sitePackages) != 1 {
		t.Fatalf("site-packages not found: %v %v", sitePackages, err)
	}
	fake, err := os.ReadFile(filepath.Join("..", "services", "testdata", "fake_realsense", "pyrealsense2.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sitePackages[0], "pyrealsense2.py"), fake, 0o644); err != nil {
		t.Fatal(err)
	}
	return project
}

func openCalibre(t *testing.T, project string) testApp {
	t.Helper()
	return openApp(t, data.Settings{Profile: domain.Calibre3DProfileID, Calibre3DProject: project})
}

func TestCameraWaitsForAReadyEnvironment(t *testing.T) {
	app := openCalibre(t, calibreProject(t))
	waitForText(t, app.window, "Prepara el entorno para usar la cámara.")
	if _, err := app.window.Element("Guardar imagen"); err == nil {
		t.Error("Guardar imagen should not be offered before the environment is ready")
	}
}

func TestSavingAPhotoAsksForItsNameAndSavesIt(t *testing.T) {
	project := readyProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, "Listo")
	click(t, app.window, "Guardar imagen")
	app.window.Type("pieza1")
	app.window.PressKey(input.KeyEnter)
	waitForText(t, app.window, "Imagen guardada: pieza1.png")
	if _, err := os.Stat(domain.PhotoPath(project, "pieza1")); err != nil {
		t.Errorf("photo not saved: %v", err)
	}
	// The line names the whole path; a long temporary folder (macOS) is cut short on screen, so only its start is checked.
	if !shows(app, "✓ Imagen guardada en ") {
		t.Errorf("the capture's last line is not shown; window shows %q", app.window.Texts())
	}
}

func TestAnInvalidPhotoNameIsExplainedAndNotSaved(t *testing.T) {
	project := readyProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, "Listo")
	click(t, app.window, "Guardar imagen")
	app.window.Type("con espacio")
	app.window.PressKey(input.KeyEnter)
	app.window.Frame()
	if !slices.Contains(app.window.Texts(), "Usa solo letras, números, - y _, sin espacios y hasta 64 caracteres.") {
		t.Errorf("no explanation shown; window shows %q", app.window.Texts())
	}
	if entries, _ := os.ReadDir(filepath.Join(project, domain.PhotosDir)); len(entries) != 0 {
		t.Errorf("something was saved: %v", entries)
	}
}

func TestCancelClosesTheNameDialog(t *testing.T) {
	app := openCalibre(t, readyProject(t))
	waitForText(t, app.window, "Listo")
	click(t, app.window, "Guardar imagen")
	click(t, app.window, "Cancelar")
	if slices.Contains(app.window.Texts(), "Cancelar") {
		t.Errorf("dialog still shown: %q", app.window.Texts())
	}
}

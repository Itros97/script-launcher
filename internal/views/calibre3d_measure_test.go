package views

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// measuringProject is a ready project whose calibre3d is the services' fake one, with the fake camera too.
func measuringProject(t *testing.T) string {
	t.Helper()
	project := readyProject(t)
	fake := filepath.Join("..", "services", "testdata", "fake_calibre3d", "calibre3d")
	for _, name := range []string{"__init__.py", "__main__.py"} {
		content, err := os.ReadFile(filepath.Join(fake, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(project, "calibre3d", name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return project
}

// shows tells whether the window shows text, even when a paragraph split it into words.
func shows(app testApp, text string) bool {
	return strings.Contains(strings.Join(app.window.Texts(), " "), text)
}

func openMeasuring(t *testing.T, project string) testApp {
	t.Helper()
	app := openCalibre(t, project)
	waitForText(t, app.window, "Listo")
	return app
}

func TestThePlaneStepSaysWhenItWasDone(t *testing.T) {
	app := openMeasuring(t, measuringProject(t))
	if !shows(app, "Deja el tapete vacío.") || !shows(app, "Haz primero el plano.") {
		t.Fatalf("no first-step hints; window shows %q", app.window.Texts())
	}
	click(t, app.window, "1. Plano")
	waitWhile(t, app, "the plane", func() bool { return !shows(app, "ruido 0,04 mm") })
	if shows(app, "Haz primero el plano.") {
		t.Error("the measure step still asks for the plane")
	}
}

func TestMeasuringAPieceShowsItsSizesAndPictures(t *testing.T) {
	project := measuringProject(t)
	app := openMeasuring(t, project)
	click(t, app.window, "3. Medir")
	app.window.Type("pieza1")
	app.window.PressKey(input.KeyEnter)
	waitForText(t, app.window, "61,23 mm")
	for _, text := range []string{"Resultado · pieza1", "40,50 mm", "19,88 mm", "62,00 mm", "Medida «pieza1»"} {
		if !slices.Contains(app.window.Texts(), text) {
			t.Errorf("missing %q; window shows %q", text, app.window.Texts())
		}
	}
	if !shows(app, "⚠ Pieza lejos del centro") {
		t.Error("calibre3d's warning is not shown")
	}
	if _, err := os.Stat(domain.StepMeasure.ResultFile(project, "pieza1", "_mapa")); err != nil {
		t.Errorf("no heat map: %v", err)
	}
}

func TestAStepClosesTheLiveCameraFirst(t *testing.T) {
	project := measuringProject(t)
	app := openMeasuring(t, project)
	click(t, app.window, "Ver cámara")
	waitWhile(t, app, "the camera", func() bool { return !app.preview.Running(project) })
	click(t, app.window, "2. Centrar")
	waitForText(t, app.window, "He parado la vista de la cámara para medir.")
	if app.preview.Running(project) {
		t.Error("the camera is still open")
	}
	waitForText(t, app.window, "Centrado (dentro de 2.0 mm). Fija la plantilla.")
}

func TestAFailedStepKeepsTheLastResult(t *testing.T) {
	app := openMeasuring(t, measuringProject(t))
	click(t, app.window, "3. Medir")
	app.window.Type("buena")
	app.window.PressKey(input.KeyEnter)
	waitForText(t, app.window, "61,23 mm")
	t.Setenv("FAKE_CALIBRE_FAIL", "1")
	click(t, app.window, "Demo sin cámara")
	waitForText(t, app.window, "No se pudo completar «Demo»: mira el detalle a la derecha.")
	if !slices.Contains(app.window.Texts(), "Resultado · buena") {
		t.Errorf("the last result is gone; window shows %q", app.window.Texts())
	}
}

// TestTheRealCalibre3DDemoInTheWindow runs the real project's demo from the window when CALIBRE3D_PROJECT names it.
func TestTheRealCalibre3DDemoInTheWindow(t *testing.T) {
	project := os.Getenv("CALIBRE3D_PROJECT")
	if project == "" {
		t.Skip("set CALIBRE3D_PROJECT to the calibre3d project's folder to run it")
	}
	app := openApp(t, data.Settings{Profile: domain.Calibre3DProfileID, Calibre3DProject: project})
	waitForText(t, app.window, "Listo")
	click(t, app.window, "Demo sin cámara")
	waitForText(t, app.window, "Medida lista: Demo")
	if !slices.Contains(app.window.Texts(), "Resultado · demo_canto") {
		t.Errorf("window shows %q", app.window.Texts())
	}
}

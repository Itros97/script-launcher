package views

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"

	"github.com/Itros97/script-launcher/internal/domain"
)

// waitWhile runs frames while condition holds, failing after a timeout with what.
func waitWhile(t *testing.T, app testApp, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s; window shows %q", what, app.window.Texts())
		}
		time.Sleep(10 * time.Millisecond)
		app.window.Frame()
	}
}

// openLiveCamera opens a ready project on Calibre3D and starts the live view until a frame is shown.
func openLiveCamera(t *testing.T) (testApp, string) {
	t.Helper()
	project := readyProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, "Listo")
	click(t, app.window, "Ver cámara")
	waitForText(t, app.window, "▸ Cámara en marcha (1280×720, 15 imágenes por segundo)")
	waitWhile(t, app, "the first frame", func() bool { return slices.Contains(app.window.Texts(), "Arrancando la cámara…") })
	return app, project
}

func TestLiveCameraShowsInTheActivityPanel(t *testing.T) {
	app, project := openLiveCamera(t)
	if !app.preview.Running(project) {
		t.Fatal("the camera should be running")
	}
	if _, err := app.window.Element("Parar cámara"); err != nil {
		t.Errorf("no way to stop the camera: %v", err)
	}
	click(t, app.window, "Parar cámara")
	waitForText(t, app.window, "■ Cámara parada")
	if _, err := app.window.Element("Ver cámara"); err != nil {
		t.Errorf("Ver cámara should be back: %v", err)
	}
}

func TestPhotoWhileLiveSavesTheFrameOnScreen(t *testing.T) {
	app, project := openLiveCamera(t)
	click(t, app.window, "Guardar imagen")
	app.window.Type("encuadre")
	app.window.PressKey(input.KeyEnter)
	waitForText(t, app.window, "Imagen guardada: encuadre.png")
	if _, err := os.Stat(domain.PhotoPath(project, "encuadre")); err != nil {
		t.Errorf("photo not saved: %v", err)
	}
	if !app.preview.Running(project) {
		t.Error("saving a photo should not stop the live camera")
	}
}

func TestLeavingCalibre3DStopsTheCamera(t *testing.T) {
	app, project := openLiveCamera(t)
	click(t, app.window, "Calibre3D")
	click(t, app.window, "General")
	waitWhile(t, app, "the camera to stop", func() bool { return app.preview.Running(project) })
}

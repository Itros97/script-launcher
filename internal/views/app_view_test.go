package views

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/akrck02/go-ui/uitest"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/profiles"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// testApp is a window over real repositories in a temporary folder and a real run service.
type testApp struct {
	window   *uitest.Window
	session  *session.Session
	scripts  data.ScriptRepository
	settings data.SettingsRepository
	preview  *services.PreviewService
}

// writeScript creates a script file in a temporary folder and returns its path.
func writeScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// openApp saves the given paths and settings, builds the window and shows its first frame.
func openApp(t *testing.T, saved data.Settings, paths ...string) testApp {
	t.Helper()
	folder := t.TempDir()
	scripts := data.NewJSONScriptRepository(data.OSFileStore{}, filepath.Join(folder, "scripts.json"), time.Now)
	settings := data.NewJSONSettingsRepository(data.OSFileStore{}, filepath.Join(folder, "settings.json"))
	if err := settings.Save(saved); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if _, err := scripts.Add(path); err != nil {
			t.Fatal(err)
		}
	}
	service := services.NewRunService(domain.DefaultRegistry(), time.Now, func() {})
	t.Cleanup(func() { _ = service.StopAll() })
	current, err := session.New(scripts, service, time.Now, "")
	if err != nil {
		t.Fatal(err)
	}
	env := services.NewPythonEnvService(service, "python3", "calibre3d")
	scriptDir := t.TempDir()
	preview := services.NewPreviewService(scriptDir, func() {})
	t.Cleanup(func() { _ = preview.Stop() })
	calibre, err := profiles.NewCalibre3D(env, services.NewCaptureService(service, preview, scriptDir), preview, settings)
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewAppView(current, profiles.DefaultRegistry(calibre), settings)
	if err != nil {
		t.Fatal(err)
	}
	window, err := uitest.New(app.Draw, uitest.Size(1100, 680))
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	return testApp{window: window, session: current, scripts: scripts, settings: settings, preview: preview}
}

// waitForText runs frames until the window shows text, failing after a timeout.
func waitForText(t *testing.T, window *uitest.Window, text string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !slices.Contains(window.Texts(), text) {
		if time.Now().After(deadline) {
			t.Fatalf("window never showed %q; it shows %q", text, window.Texts())
		}
		time.Sleep(10 * time.Millisecond)
		window.Frame()
	}
}

// click presses an element by its label and runs a frame, failing the test when it is not there.
func click(t *testing.T, window *uitest.Window, label string) {
	t.Helper()
	if err := window.Click(label); err != nil {
		t.Fatal(err)
	}
	window.Frame()
}

func TestShowsAHintWhenNoScriptIsSaved(t *testing.T) {
	app := openApp(t, data.DefaultSettings())
	if !slices.Contains(app.window.Texts(), "Aún no hay scripts.") {
		t.Errorf("got %q", app.window.Texts())
	}
}

func TestRunButtonShowsTheScriptsOutput(t *testing.T) {
	app := openApp(t, data.DefaultSettings(), writeScript(t, "hello.sh", "echo hello from the script\n"))
	click(t, app.window, "Ejecutar")
	waitForText(t, app.window, "hello from the script")
	app.window.Frame()
	if _, err := app.window.Element("Ejecutar"); err != nil {
		t.Errorf("Run button should be back after the script ends: %v", err)
	}
}

func TestRemoveTakesTheScriptOutOfTheSavedList(t *testing.T) {
	app := openApp(t, data.DefaultSettings(), writeScript(t, "bye.sh", "true\n"))
	click(t, app.window, "Quitar de la lista")
	scripts, err := app.scripts.List()
	if err != nil || len(scripts) != 0 {
		t.Fatalf("got %v, %v", scripts, err)
	}
	waitForText(t, app.window, "Aún no hay scripts.")
}

func TestMissingFileHasNoRunButton(t *testing.T) {
	app := openApp(t, data.DefaultSettings(), filepath.Join(t.TempDir(), "gone.sh"))
	if _, err := app.window.Element("Ejecutar"); err == nil {
		t.Error("a missing file should not offer Run")
	}
}

func TestAddingTheSameFileTwiceKeepsOneRow(t *testing.T) {
	path := writeScript(t, "once.sh", "true\n")
	app := openApp(t, data.DefaultSettings(), path)
	app.session.Add(path)
	notice, _ := app.session.CurrentNotice()
	if len(app.session.Scripts()) != 1 || notice.Text != "Ese script ya está en la lista" {
		t.Errorf("got %d scripts, notice %q", len(app.session.Scripts()), notice.Text)
	}
}

func TestAddScriptOpensTheFilePicker(t *testing.T) {
	app := openApp(t, data.DefaultSettings())
	click(t, app.window, "Añadir script")
	if !slices.Contains(app.window.Texts(), "Añadir un script") {
		t.Errorf("picker not shown; window shows %q", app.window.Texts())
	}
}

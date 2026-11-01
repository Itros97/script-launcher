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
	calibre, err := profiles.NewCalibre3D(profiles.Calibre3DServices{
		Env: env, Camera: services.NewCaptureService(service, preview, scriptDir), Preview: preview,
		Calibre: services.NewCalibreService(service, preview), Settings: settings,
	})
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

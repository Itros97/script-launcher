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
)

// writeScript creates a script file in a temporary folder and returns its path.
func writeScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// openApp saves the given paths, builds the view over a real run service and shows its first frame.
func openApp(t *testing.T, paths ...string) (*uitest.Window, *AppView, data.ScriptRepository) {
	t.Helper()
	repository := data.NewJSONScriptRepository(data.OSFileStore{}, filepath.Join(t.TempDir(), "scripts.json"), time.Now)
	for _, path := range paths {
		if _, err := repository.Add(path); err != nil {
			t.Fatal(err)
		}
	}
	service := services.NewRunService(domain.DefaultRegistry(), time.Now, func() {})
	t.Cleanup(func() { _ = service.StopAll() })
	app, err := NewAppView(repository, service, time.Now, "")
	if err != nil {
		t.Fatal(err)
	}
	window, err := uitest.New(app.Draw, uitest.Size(1100, 680))
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	return window, app, repository
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

func TestShowsAHintWhenNoScriptIsSaved(t *testing.T) {
	window, _, _ := openApp(t)
	if !slices.Contains(window.Texts(), "No scripts yet.") {
		t.Errorf("got %q", window.Texts())
	}
}

func TestRunButtonShowsTheScriptsOutput(t *testing.T) {
	path := writeScript(t, "hello.sh", "echo hello from the script\n")
	window, _, _ := openApp(t, path)
	if err := window.Click("Run"); err != nil {
		t.Fatal(err)
	}
	waitForText(t, window, "hello from the script")
	window.Frame()
	if _, err := window.Element("Run"); err != nil {
		t.Errorf("Run button should be back after the script ends: %v", err)
	}
}

func TestRemoveTakesTheScriptOutOfTheSavedList(t *testing.T) {
	path := writeScript(t, "bye.sh", "true\n")
	window, _, repository := openApp(t, path)
	if err := window.Click("Remove from the list"); err != nil {
		t.Fatal(err)
	}
	window.Frame()
	scripts, err := repository.List()
	if err != nil || len(scripts) != 0 {
		t.Fatalf("got %v, %v", scripts, err)
	}
	waitForText(t, window, "No scripts yet.")
}

func TestMissingFileHasNoRunButton(t *testing.T) {
	window, _, _ := openApp(t, filepath.Join(t.TempDir(), "gone.sh"))
	if _, err := window.Element("Run"); err == nil {
		t.Error("a missing file should not offer Run")
	}
}

func TestAddingTheSameFileTwiceKeepsOneRow(t *testing.T) {
	path := writeScript(t, "once.sh", "true\n")
	_, app, _ := openApp(t, path)
	app.addScript(path)
	if len(app.scripts) != 1 || app.notice != "That script is already in the list" {
		t.Errorf("got %d scripts, notice %q", len(app.scripts), app.notice)
	}
}

func TestShortPathWritesHomeAsTilde(t *testing.T) {
	cases := map[string]string{"/home/user/bin": "~/bin", "/home/user": "~", "/home/username": "/home/username", "/opt/x": "/opt/x"}
	for path, want := range cases {
		if got := shortPath(path, "/home/user"); got != want {
			t.Errorf("shortPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestFormatDurationKeepsShortRunsReadable(t *testing.T) {
	cases := map[time.Duration]string{42 * time.Millisecond: "42ms", 1540 * time.Millisecond: "1.5s", 90 * time.Second: "1m30s"}
	for duration, want := range cases {
		if got := formatDuration(duration); got != want {
			t.Errorf("formatDuration(%s) = %q, want %q", duration, got, want)
		}
	}
}

func TestAddScriptOpensTheFilePicker(t *testing.T) {
	window, _, _ := openApp(t)
	if err := window.Click("Add script"); err != nil {
		t.Fatal(err)
	}
	window.Frame()
	if !slices.Contains(window.Texts(), "Add a script") {
		t.Errorf("picker not shown; window shows %q", window.Texts())
	}
}

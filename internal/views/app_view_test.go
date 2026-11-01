package views

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/Itros97/script-launcher/internal/data"
)

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

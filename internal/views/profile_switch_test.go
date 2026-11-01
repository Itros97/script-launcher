package views

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// calibreProject creates a calibre3d project folder with an importable module and no requirements.
func calibreProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "calibre3d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calibre3d", "__init__.py"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestChoosingAProfileChangesTheLayoutAndIsSaved(t *testing.T) {
	app := openApp(t, data.DefaultSettings())
	click(t, app.window, "General")
	click(t, app.window, "Calibre3D")
	texts := app.window.Texts()
	if !slices.Contains(texts, "Sin proyecto") || slices.Contains(texts, "SCRIPTS GUARDADOS") {
		t.Errorf("Calibre3D layout not shown; window shows %q", texts)
	}
	saved, err := app.settings.Load()
	if err != nil || saved.Profile != domain.Calibre3DProfileID {
		t.Errorf("got %+v, %v", saved, err)
	}
}

func TestOpensOnTheProfileSavedLastTime(t *testing.T) {
	app := openApp(t, data.Settings{Profile: domain.Calibre3DProfileID})
	if !slices.Contains(app.window.Texts(), "Sin proyecto") {
		t.Errorf("got %q", app.window.Texts())
	}
}

func TestUnknownSavedProfileOpensTheGeneralLayout(t *testing.T) {
	app := openApp(t, data.Settings{Profile: "gone"})
	if !slices.Contains(app.window.Texts(), "SCRIPTS GUARDADOS") {
		t.Errorf("got %q", app.window.Texts())
	}
}

func TestARunningScriptKeepsRunningAcrossProfiles(t *testing.T) {
	path := writeScript(t, "long.sh", "echo first line\nsleep 30\n")
	app := openApp(t, data.DefaultSettings(), path)
	click(t, app.window, "Ejecutar")
	waitForText(t, app.window, "first line")
	click(t, app.window, "General")
	click(t, app.window, "Calibre3D")
	click(t, app.window, "Calibre3D")
	click(t, app.window, "General")
	run, ok := app.session.Latest(domain.ScriptID(path))
	if !ok || !run.Running() {
		t.Fatal("the script stopped when the profile changed")
	}
	waitForText(t, app.window, "first line")
}

func TestCalibre3DPreparesTheEnvironmentInOneClick(t *testing.T) {
	project := calibreProject(t)
	app := openApp(t, data.Settings{Profile: domain.Calibre3DProfileID, Calibre3DProject: project})
	waitForText(t, app.window, "Sin entorno")
	click(t, app.window, "Preparar entorno")
	waitForText(t, app.window, "Listo")
	if _, err := os.Stat(filepath.Join(project, domain.VenvPython)); err != nil {
		t.Errorf("no venv was created: %v", err)
	}
	waitForText(t, app.window, "✓ Entorno preparado")
}

func TestCalibre3DOffersAFolderPicker(t *testing.T) {
	app := openApp(t, data.Settings{Profile: domain.Calibre3DProfileID})
	click(t, app.window, "Elegir carpeta")
	if !slices.Contains(app.window.Texts(), "Carpeta de calibre3d") {
		t.Errorf("picker not shown; window shows %q", app.window.Texts())
	}
}

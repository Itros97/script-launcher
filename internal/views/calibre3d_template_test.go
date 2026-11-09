package views

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/templates"
)

// The launcher picks and offers the template of the computer the tests run on; other is the other one.
var (
	this, other         = platforms()
	missingRequirements = "Sin requirements.txt: se usará la plantilla de " + this.Name() + "."
)

func platforms() (domain.Platform, domain.Platform) {
	if this := domain.PlatformFor(runtime.GOOS); this == domain.MacOS {
		return this, domain.Linux
	}
	return domain.Linux, domain.MacOS
}

func TestSetupWithoutRequirementsInstallsThisComputersTemplate(t *testing.T) {
	project := calibreProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, missingRequirements)
	click(t, app.window, "Preparar entorno")
	waitForText(t, app.window, "Preparando…")
	content, err := os.ReadFile(filepath.Join(project, this.RequirementsFile()))
	if err != nil || !bytes.Equal(content, templates.Requirements(this)) {
		t.Errorf("%s is not the template: %v", this.RequirementsFile(), err)
	}
	if slices.Contains(app.window.Texts(), missingRequirements) {
		t.Error("the missing requirements hint is still shown")
	}
}

func TestEachTemplateButtonSavesItsOwnFile(t *testing.T) {
	project := calibreProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, missingRequirements)
	click(t, app.window, "Plantilla "+other.Name())
	waitForText(t, app.window, "Plantilla guardada: "+other.RequirementsFile())
	if content, _ := os.ReadFile(filepath.Join(project, other.RequirementsFile())); !bytes.Equal(content, templates.Requirements(other)) {
		t.Errorf("%s is not its template: %q", other.RequirementsFile(), content)
	}
	waitForText(t, app.window, missingRequirements) // the other computer's file is not installed here
	click(t, app.window, "Plantilla "+this.Name())
	waitForText(t, app.window, "Plantilla guardada: "+this.RequirementsFile())
	waitForText(t, app.window, "Se instalará "+this.RequirementsFile()+".")
}

func TestThisComputersFileWinsOverRequirementsTxt(t *testing.T) {
	project := calibreProject(t)
	if err := os.WriteFile(filepath.Join(project, domain.RequirementsFile), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := openCalibre(t, project)
	waitForText(t, app.window, "Se instalará requirements.txt.")
	click(t, app.window, "Plantilla "+this.Name())
	waitForText(t, app.window, "Se instalará "+this.RequirementsFile()+".")
	if content, _ := os.ReadFile(filepath.Join(project, domain.RequirementsFile)); string(content) != "mine\n" {
		t.Errorf("requirements.txt was changed: %q", content)
	}
}

func TestTemplateButtonNeverOverwrites(t *testing.T) {
	project := calibreProject(t)
	if err := os.WriteFile(filepath.Join(project, domain.MacOS.RequirementsFile()), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := openCalibre(t, project)
	waitForText(t, app.window, "Sin entorno")
	click(t, app.window, "Plantilla macOS")
	waitForText(t, app.window, "Esa plantilla ya está en la carpeta del proyecto: borra el archivo si quieres guardarla otra vez.")
	if content, _ := os.ReadFile(filepath.Join(project, domain.MacOS.RequirementsFile())); string(content) != "mine\n" {
		t.Errorf("requirements-macos.txt was changed: %q", content)
	}
}

func TestIncompleteEnvironmentSaysWhy(t *testing.T) {
	project := readyProject(t)
	if err := os.RemoveAll(filepath.Join(project, "calibre3d")); err != nil {
		t.Fatal(err)
	}
	app := openCalibre(t, project)
	waitForText(t, app.window, "Incompleto")
	shown := strings.Join(app.window.Texts(), " ")
	if !strings.Contains(shown, "En esta carpeta no está el programa calibre3d. Elige la carpeta del proyecto (la que tiene prueba.sh).") {
		t.Errorf("no reason shown; window shows %q", shown)
	}
}

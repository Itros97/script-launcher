package views

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/templates"
)

const missingRequirements = "Sin requirements.txt: se usará la plantilla."

func TestSetupWithoutRequirementsInstallsTheTemplate(t *testing.T) {
	project := calibreProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, missingRequirements)
	click(t, app.window, "Preparar entorno")
	waitForText(t, app.window, "Preparando…")
	content, err := os.ReadFile(filepath.Join(project, domain.RequirementsFile))
	if err != nil || !bytes.Equal(content, templates.Requirements) {
		t.Errorf("requirements.txt is not the template: %v", err)
	}
	if slices.Contains(app.window.Texts(), missingRequirements) {
		t.Error("the missing requirements hint is still shown")
	}
}

func TestTemplateButtonWritesTheRequirementsAndClearsTheHint(t *testing.T) {
	project := calibreProject(t)
	app := openCalibre(t, project)
	waitForText(t, app.window, missingRequirements)
	click(t, app.window, "Descargar plantilla de requirements.txt")
	waitForText(t, app.window, "Plantilla guardada: requirements.txt")
	if _, err := os.Stat(filepath.Join(project, domain.RequirementsFile)); err != nil {
		t.Errorf("requirements.txt not written: %v", err)
	}
	if slices.Contains(app.window.Texts(), missingRequirements) {
		t.Error("the missing requirements hint is still shown")
	}
}

func TestTemplateIsSavedBesideExistingRequirements(t *testing.T) {
	project := calibreProject(t)
	if err := os.WriteFile(filepath.Join(project, domain.RequirementsFile), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := openCalibre(t, project)
	waitForText(t, app.window, "Sin entorno")
	click(t, app.window, "Descargar plantilla de requirements.txt")
	waitForText(t, app.window, "Plantilla guardada: requirements.plantilla.txt")
	if content, _ := os.ReadFile(filepath.Join(project, domain.RequirementsFile)); string(content) != "mine\n" {
		t.Errorf("requirements.txt was changed: %q", content)
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

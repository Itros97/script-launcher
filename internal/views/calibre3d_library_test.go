package views

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// saveResult leaves a calibre3d result for a piece in the project, with its sizes and the day of November it was measured.
func saveResult(t *testing.T, project, name string, day int, length, width, thickness float64) {
	t.Helper()
	folder := domain.ResultsDir(project)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf(`{"nombre": %q, "fecha": "2026-11-%02dT10:00:00", "resultado": {"largo_mm": %g, "ancho_mm": %g, "espesor_mm": %g, "feret_max_mm": 1, "aviso": []}}`,
		name, day, length, width, thickness)
	if err := os.WriteFile(filepath.Join(folder, name+".json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// openLibrary opens the Calibre3D profile on the project, with exports going to exports, and shows its library.
func openLibrary(t *testing.T, project, exports string) testApp {
	t.Helper()
	app := openApp(t, data.Settings{Profile: domain.Calibre3DProfileID, Calibre3DProject: project, Calibre3DExportFolder: exports})
	click(t, app.window, "Abrir biblioteca")
	return app
}

// clickIn clicks the control with this label in a measurement's row.
func clickIn(t *testing.T, app testApp, row, label string) {
	t.Helper()
	element, err := app.window.ElementIn(row, label)
	if err != nil {
		t.Fatal(err)
	}
	app.window.ClickAt(element.Rect.X+element.Rect.Width/2, element.Rect.Y+element.Rect.Height/2)
	app.window.Frame()
}

// threePieces is a project with three measurements: tapa (newest), base and tapa-grande.
func threePieces(t *testing.T) string {
	project := calibreProject(t)
	saveResult(t, project, "base", 2, 30, 20, 10)
	saveResult(t, project, "tapa", 3, 60, 40, 5)
	saveResult(t, project, "tapa-grande", 1, 90, 60, 5)
	return project
}

func TestAnEmptyLibrarySaysSo(t *testing.T) {
	app := openLibrary(t, calibreProject(t), "")
	if !shows(app, "Aún no hay mediciones en este proyecto.") || !shows(app, "0 de 0 mediciones") {
		t.Errorf("window shows %q", app.window.Texts())
	}
}

func TestTheLibraryListsTheMeasurementsNewestFirst(t *testing.T) {
	app := openLibrary(t, threePieces(t), "")
	texts := app.window.Texts()
	order := []int{slices.Index(texts, "tapa"), slices.Index(texts, "base"), slices.Index(texts, "tapa-grande")}
	if order[0] < 0 || !slices.IsSorted(order) || !shows(app, "3 de 3 mediciones") {
		t.Errorf("order %v; window shows %q", order, texts)
	}
	click(t, app.window, "Volver")
	if !shows(app, "Actividad") {
		t.Error("Volver did not go back to the activity")
	}
}

func TestTheLibraryShowsTwentyMeasurementsPerPage(t *testing.T) {
	project := calibreProject(t)
	for index := range 25 {
		saveResult(t, project, fmt.Sprintf("pieza%02d", index), 1+index%28, 10, 10, 10)
	}
	app := openLibrary(t, project, "")
	if !shows(app, "Página 1 de 2") {
		t.Fatalf("window shows %q", app.window.Texts())
	}
	click(t, app.window, "Siguiente")
	rows := 0
	for _, text := range app.window.Texts() {
		if strings.HasPrefix(text, "pieza") {
			rows++
		}
	}
	if !shows(app, "Página 2 de 2") || rows != 5 {
		t.Errorf("%d rows; window shows %q", rows, app.window.Texts())
	}
}

func TestTheNameFilterLeavesOnlyTheMatchingMeasurements(t *testing.T) {
	app := openLibrary(t, threePieces(t), "")
	click(t, app.window, "Buscar por nombre")
	app.window.Type("TAPA")
	app.window.Frame()
	if !shows(app, "2 de 3 mediciones") || slices.Contains(app.window.Texts(), "base") {
		t.Errorf("window shows %q", app.window.Texts())
	}
}

func TestTheAxisFiltersKeepTheMeasurementsInsideTheirRange(t *testing.T) {
	app := openLibrary(t, threePieces(t), "")
	clickIn(t, app, "Largo", "mín")
	app.window.Type("50")
	clickIn(t, app, "Largo", "máx")
	app.window.Type("60,0")
	app.window.Frame()
	if !shows(app, "1 de 3 mediciones") || !slices.Contains(app.window.Texts(), "tapa") {
		t.Errorf("window shows %q", app.window.Texts())
	}
	clickIn(t, app, "Espesor", "mín")
	app.window.Type("doce")
	app.window.Frame()
	if !shows(app, "Escribe los límites en milímetros") || !shows(app, "1 de 3 mediciones") {
		t.Errorf("an invalid bound: window shows %q", app.window.Texts())
	}
	click(t, app.window, "Limpiar")
	if !shows(app, "3 de 3 mediciones") {
		t.Errorf("Limpiar: window shows %q", app.window.Texts())
	}
}

func TestDeletingAMeasurementAsksFirstAndRemovesItsFiles(t *testing.T) {
	project := threePieces(t)
	app := openLibrary(t, project, "")
	clickIn(t, app, "base", "##Borrar")
	if !shows(app, "¿Borrar «base»?") {
		t.Fatalf("no confirmation; window shows %q", app.window.Texts())
	}
	click(t, app.window, "Cancelar")
	if _, err := os.Stat(domain.PieceFiles(project, "base")[0]); err != nil {
		t.Fatalf("Cancelar deleted it: %v", err)
	}
	clickIn(t, app, "base", "##Borrar")
	click(t, app.window, "Borrar")
	if _, err := os.Stat(domain.PieceFiles(project, "base")[0]); !os.IsNotExist(err) {
		t.Errorf("base is still on disk: %v", err)
	}
	if !shows(app, "2 de 2 mediciones") {
		t.Errorf("window shows %q", app.window.Texts())
	}
}

func TestExportingTheWholePageWritesACSVAndRemembersTheFolder(t *testing.T) {
	exports := t.TempDir()
	app := openLibrary(t, threePieces(t), exports)
	click(t, app.window, "Exportar CSV")
	click(t, app.window, "Esta página (3)")
	click(t, app.window, "Choose")
	files, _ := filepath.Glob(filepath.Join(exports, "mediciones-*.csv"))
	if len(files) != 1 {
		t.Fatalf("got %v; window shows %q", files, app.window.Texts())
	}
	content, _ := os.ReadFile(files[0])
	if lines := strings.Count(string(content), "\n"); lines != 4 {
		t.Errorf("%d lines in\n%s", lines, content)
	}
	saved, err := app.settings.Load()
	if err != nil || saved.Calibre3DExportFolder != exports {
		t.Errorf("saved %+v, %v", saved, err)
	}
}

func TestSelectingTheWholePageDeletesEveryMeasurementOnIt(t *testing.T) {
	app := openLibrary(t, threePieces(t), "")
	click(t, app.window, "Toda la página")
	click(t, app.window, "Borrar seleccionadas (3)")
	if !shows(app, "¿Borrar 3 mediciones?") {
		t.Fatalf("window shows %q", app.window.Texts())
	}
	click(t, app.window, "Borrar")
	if !shows(app, "Aún no hay mediciones en este proyecto.") {
		t.Errorf("window shows %q", app.window.Texts())
	}
}

func TestExportingTheSelectionWritesOnlyTheSelectedMeasurements(t *testing.T) {
	exports := t.TempDir()
	app := openLibrary(t, threePieces(t), exports)
	clickIn(t, app, "base", "##Seleccionar")
	click(t, app.window, "Exportar CSV")
	click(t, app.window, "Seleccionadas (1)")
	click(t, app.window, "Choose")
	files, _ := filepath.Glob(filepath.Join(exports, "mediciones-*.csv"))
	if len(files) != 1 {
		t.Fatalf("got %v", files)
	}
	content, _ := os.ReadFile(files[0])
	if strings.Count(string(content), "\n") != 2 || !strings.Contains(string(content), "\nbase;") {
		t.Errorf("got\n%s", content)
	}
}

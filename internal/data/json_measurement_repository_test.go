package data

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

// writeResult leaves a calibre3d result file in the project, as calibre3d would.
func writeResult(t *testing.T, project, name, content string) {
	t.Helper()
	folder := domain.ResultsDir(project)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func result(name, date string) string {
	return `{"nombre": "` + name + `", "fecha": "` + date + `", "resultado": {"largo_mm": 10, "ancho_mm": 5, "espesor_mm": 2, "feret_max_mm": 11, "aviso": []}}`
}

func TestListReadsThePiecesNewestFirstSkippingTheCentringAndCorruptFiles(t *testing.T) {
	project := t.TempDir()
	writeResult(t, project, "vieja.json", result("vieja", "2026-11-01T10:00:00"))
	writeResult(t, project, "nueva.json", result("nueva", "2026-11-20T10:00:00"))
	writeResult(t, project, "centrado.json", `{"resultado": {}}`)
	writeResult(t, project, "rota.json", "{oops")
	writeResult(t, project, "nueva_mapa.png", "png")
	measurements, unreadable, err := JSONMeasurementRepository{}.List(project)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, measurement := range measurements {
		names = append(names, measurement.Name)
	}
	if !slices.Equal(names, []string{"nueva", "vieja"}) || unreadable != 1 {
		t.Errorf("got %v and %d unreadable", names, unreadable)
	}
}

func TestListOfAProjectWithoutResultsIsEmpty(t *testing.T) {
	for _, project := range []string{"", t.TempDir()} {
		measurements, unreadable, err := JSONMeasurementRepository{}.List(project)
		if err != nil || len(measurements) != 0 || unreadable != 0 {
			t.Errorf("%q: got %v, %d, %v", project, measurements, unreadable, err)
		}
	}
}

func TestDeleteRemovesTheResultAndItsPictures(t *testing.T) {
	project := t.TempDir()
	writeResult(t, project, "pieza.json", result("pieza", "2026-11-20T10:00:00"))
	writeResult(t, project, "pieza_mapa.png", "png")
	writeResult(t, project, "otra.json", result("otra", "2026-11-20T10:00:00"))
	if err := (JSONMeasurementRepository{}).Delete(project, "pieza"); err != nil {
		t.Fatal(err)
	}
	for _, path := range domain.PieceFiles(project, "pieza") {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s is still there", path)
		}
	}
	if _, err := os.Stat(filepath.Join(domain.ResultsDir(project), "otra.json")); err != nil {
		t.Errorf("another measurement was deleted: %v", err)
	}
}

func TestDeletingAMissingMeasurementIsAnError(t *testing.T) {
	if err := (JSONMeasurementRepository{}).Delete(t.TempDir(), "nada"); err == nil {
		t.Error("expected an error")
	}
}

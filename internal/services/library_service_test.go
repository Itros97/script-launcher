package services

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// fakeMeasurements is a MeasurementRepository in memory that fails to delete the names in broken.
type fakeMeasurements struct {
	deleted []string
	broken  []string
}

func (fake *fakeMeasurements) List(string) ([]domain.Measurement, int, error) { return nil, 0, nil }

func (fake *fakeMeasurements) Delete(_, name string) error {
	if slices.Contains(fake.broken, name) {
		return errors.New("cannot delete " + name)
	}
	fake.deleted = append(fake.deleted, name)
	return nil
}

func fixedClock() time.Time { return time.Date(2026, 11, 24, 15, 30, 12, 0, time.Local) }

func TestExportWritesACSVNamedAfterTheMoment(t *testing.T) {
	folder := t.TempDir()
	service := NewLibraryService(&fakeMeasurements{}, fixedClock)
	path, err := service.Export(folder, []domain.Measurement{{Name: "pieza1", Length: 61.23}})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(folder, "mediciones-2026-11-24_153012.csv") {
		t.Errorf("path %s", path)
	}
	content, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(content), "pieza1;;61,2300;") {
		t.Errorf("got %q, %v", content, err)
	}
}

func TestExportNeverOverwritesAFile(t *testing.T) {
	folder := t.TempDir()
	service := NewLibraryService(&fakeMeasurements{}, fixedClock)
	measurements := []domain.Measurement{{Name: "a"}}
	if _, err := service.Export(folder, measurements); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Export(folder, measurements); err == nil {
		t.Error("the second export in the same second overwrote the first")
	}
}

func TestExportingNothingIsAnError(t *testing.T) {
	if _, err := NewLibraryService(&fakeMeasurements{}, fixedClock).Export(t.TempDir(), nil); err == nil {
		t.Error("expected an error")
	}
}

func TestDeleteGoesOnAfterAFailure(t *testing.T) {
	fake := &fakeMeasurements{broken: []string{"b"}}
	deleted, err := NewLibraryService(fake, fixedClock).Delete("/project", []string{"a", "b", "c"})
	if deleted != 2 || err == nil || !slices.Equal(fake.deleted, []string{"a", "c"}) {
		t.Errorf("deleted %d %v, err %v", deleted, fake.deleted, err)
	}
}

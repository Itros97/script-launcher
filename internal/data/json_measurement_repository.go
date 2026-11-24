package data

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Itros97/script-launcher/internal/domain"
)

// JSONMeasurementRepository reads calibre3d's result files from datos/resultados on disk.
type JSONMeasurementRepository struct{}

// List reads every piece's result file; a missing folder is an empty library.
func (JSONMeasurementRepository) List(project string) ([]domain.Measurement, int, error) {
	if project == "" {
		return nil, 0, nil
	}
	folder := domain.ResultsDir(project)
	entries, err := os.ReadDir(folder)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("reading the measurements in %s: %w", folder, err)
	}
	measurements, unreadable := []domain.Measurement{}, 0
	for _, entry := range entries {
		if entry.IsDir() || !domain.IsPieceResult(entry.Name()) {
			continue
		}
		measurement, err := readMeasurement(filepath.Join(folder, entry.Name()))
		if err != nil {
			unreadable++
			continue
		}
		measurements = append(measurements, measurement)
	}
	slices.SortFunc(measurements, newestFirst)
	return measurements, unreadable, nil
}

// Delete removes the result file, which must exist, and the pictures, which may not.
func (JSONMeasurementRepository) Delete(project, name string) error {
	for index, path := range domain.PieceFiles(project, name) {
		err := os.Remove(path)
		if errors.Is(err, fs.ErrNotExist) && index > 0 {
			continue
		}
		if err != nil {
			return fmt.Errorf("deleting the measurement %s: %w", name, err)
		}
	}
	return nil
}

// readMeasurement reads one result file; it is named after the file, since that is what finds it again.
func readMeasurement(path string) (domain.Measurement, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Measurement{}, err
	}
	measurement, err := domain.ParseMeasurement(content)
	measurement.Name = strings.TrimSuffix(filepath.Base(path), ".json")
	return measurement, err
}

// newestFirst orders by date, newest first, then by name.
func newestFirst(a, b domain.Measurement) int {
	return cmp.Or(b.Date.Compare(a.Date), cmp.Compare(a.Name, b.Name))
}

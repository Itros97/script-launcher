package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// exportName is how an exported CSV is named, after the moment it was written.
const exportName = "mediciones-2006-01-02_150405.csv"

// LibraryService lists, deletes and exports the measurements saved in a calibre3d project.
type LibraryService struct {
	repository data.MeasurementRepository
	now        func() time.Time
}

// NewLibraryService builds a service over repository; now names the exported files.
func NewLibraryService(repository data.MeasurementRepository, now func() time.Time) *LibraryService {
	return &LibraryService{repository: repository, now: now}
}

// List returns the project's measurements, newest first, and how many files could not be read.
func (service *LibraryService) List(project string) ([]domain.Measurement, int, error) {
	return service.repository.List(project)
}

// Delete removes every named measurement, going on after a failure; it returns how many it
// deleted and the failures joined.
func (service *LibraryService) Delete(project string, names []string) (int, error) {
	deleted, failures := 0, []error{}
	for _, name := range names {
		if err := service.repository.Delete(project, name); err != nil {
			failures = append(failures, err)
			continue
		}
		deleted++
	}
	return deleted, errors.Join(failures...)
}

// Export writes the measurements as a new CSV file in folder and returns its path.
func (service *LibraryService) Export(folder string, measurements []domain.Measurement) (string, error) {
	if len(measurements) == 0 {
		return "", errors.New("there are no measurements to export")
	}
	path := filepath.Join(folder, service.now().Format(exportName))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("creating the CSV: %w", err)
	}
	writeErr := domain.WriteMeasurementsCSV(file, measurements)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	return path, nil
}

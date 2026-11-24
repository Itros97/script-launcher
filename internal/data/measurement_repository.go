package data

import "github.com/Itros97/script-launcher/internal/domain"

// MeasurementRepository reads and deletes the measurements calibre3d saved in a project.
type MeasurementRepository interface {
	// List returns the project's measurements, newest first, and how many result files could not be read.
	List(project string) (measurements []domain.Measurement, unreadable int, err error)
	// Delete removes a measurement's result and its pictures.
	Delete(project, name string) error
}

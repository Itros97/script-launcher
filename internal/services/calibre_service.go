package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// cameraReleaseTimeout bounds how long a step waits for the live view to let go of the camera.
const cameraReleaseTimeout = 5 * time.Second

// CalibreService runs the calibre3d project's commands with its own Python and reads what they leave.
type CalibreService struct {
	runs    *RunService
	preview *PreviewService
}

// NewCalibreService builds a service that closes preview before each command, since calibre3d needs the camera alone.
func NewCalibreService(runs *RunService, preview *PreviewService) *CalibreService {
	return &CalibreService{runs: runs, preview: preview}
}

// Validate checks that name can name a new measurement in the project.
func (service *CalibreService) Validate(project, name string) error {
	if err := domain.ValidatePhotoAlias(name); err != nil {
		return err
	}
	if _, err := os.Stat(domain.StepMeasure.ResultFile(project, name, "")); err == nil {
		return fmt.Errorf("%w: %s", domain.ErrAliasTaken, name)
	}
	return nil
}

// Run starts a step in the project, closing the live view first; it tells whether a view was closed.
func (service *CalibreService) Run(project string, step domain.CalibreStep, name string) (*Run, bool, error) {
	if step == domain.StepMeasure {
		if err := service.Validate(project, name); err != nil {
			return nil, false, err
		}
	}
	closed, err := service.preview.StopAndWait(cameraReleaseTimeout)
	if err != nil {
		return nil, closed, err
	}
	command := Command{
		Name: step.Title(name), Runner: domain.CalibreModule,
		Program: filepath.Join(project, domain.VenvPython), Args: step.Args(name), Dir: project,
	}
	run, err := service.runs.StartCommand(calibreRunID(project), command)
	return run, closed, err
}

// Latest returns the last step started in the project in this session.
func (service *CalibreService) Latest(project string) (*Run, bool) {
	return service.runs.Latest(calibreRunID(project))
}

// LoadMeasurement reads the result a measuring step left.
func (service *CalibreService) LoadMeasurement(project string, step domain.CalibreStep, name string) (domain.Measurement, error) {
	content, err := os.ReadFile(step.ResultFile(project, name, ""))
	if err != nil {
		return domain.Measurement{}, fmt.Errorf("the measurement has no result file: %w", err)
	}
	return domain.ParseMeasurement(content)
}

// LoadPlane reads the project's reference plane; ok is false when none was saved yet.
func (service *CalibreService) LoadPlane(project string) (plane domain.PlaneInfo, ok bool, err error) {
	content, err := os.ReadFile(domain.PlaneFile(project))
	if os.IsNotExist(err) {
		return domain.PlaneInfo{}, false, nil
	}
	if err != nil {
		return domain.PlaneInfo{}, false, err
	}
	plane, err = domain.ParsePlane(content)
	return plane, err == nil, err
}

func calibreRunID(project string) string { return "calibre:" + filepath.Clean(project) }

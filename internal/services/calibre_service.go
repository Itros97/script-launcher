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
// A pattern needs its real sizes, so it is started with RunPattern instead.
func (service *CalibreService) Run(project string, step domain.CalibreStep, name string) (*Run, bool, error) {
	if step == domain.StepPattern {
		return nil, false, fmt.Errorf("a pattern needs its real sizes: start it with RunPattern")
	}
	if step == domain.StepMeasure {
		if err := service.Validate(project, name); err != nil {
			return nil, false, err
		}
	}
	return service.start(project, step.Title(name), step.Args(name))
}

// RunPattern measures a piece of known size and adds it to the project's calibration, closing the live view first.
func (service *CalibreService) RunPattern(project string, pattern domain.Pattern) (*Run, bool, error) {
	return service.start(project, domain.StepPattern.Title(pattern.Name), pattern.Args())
}

// start closes the live view, since calibre3d needs the camera alone, and runs calibre3d with these arguments.
func (service *CalibreService) start(project, title string, args []string) (*Run, bool, error) {
	closed, err := service.preview.StopAndWait(cameraReleaseTimeout)
	if err != nil {
		return nil, closed, err
	}
	command := Command{
		Name: title, Runner: domain.CalibreModule,
		Program: filepath.Join(project, domain.VenvPython), Args: args, Dir: project,
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

// LoadCalibration reads the project's calibration; ok is false while no pattern was measured.
func (service *CalibreService) LoadCalibration(project string) (calibration domain.Calibration, ok bool, err error) {
	content, err := os.ReadFile(domain.CalibrationFile(project))
	if os.IsNotExist(err) {
		return domain.Calibration{}, false, nil
	}
	if err != nil {
		return domain.Calibration{}, false, err
	}
	calibration, err = domain.ParseCalibration(content)
	return calibration, err == nil, err
}

func calibreRunID(project string) string { return "calibre:" + filepath.Clean(project) }

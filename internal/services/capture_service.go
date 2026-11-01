package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/capture"
)

// ErrCameraStarting is returned when a photo is asked of the live view before its first frame.
var ErrCameraStarting = errors.New("the camera view has no frame yet")

// CaptureService saves photos from a RealSense camera: from the live view's frame while it is open,
// or with a short capture run of the project's own Python otherwise.
type CaptureService struct {
	runs      *RunService
	preview   *PreviewService
	scriptDir string
}

// NewCaptureService builds a service that writes its capture script into scriptDir before each run.
func NewCaptureService(runs *RunService, preview *PreviewService, scriptDir string) *CaptureService {
	return &CaptureService{runs: runs, preview: preview, scriptDir: scriptDir}
}

// Live tells whether photos of the project come from the open live view.
func (service *CaptureService) Live(project string) bool { return service.preview.Running(project) }

// SaveLive saves the frame the live view shows now, and returns where.
func (service *CaptureService) SaveLive(project, alias string) (string, error) {
	if err := service.Validate(project, alias); err != nil {
		return "", err
	}
	var frame domain.Frame
	copied := service.preview.WithLatest(project, func(latest domain.Frame, _ uint64) {
		frame = domain.Frame{Width: latest.Width, Height: latest.Height, RGB: append([]byte(nil), latest.RGB...)}
	})
	if !copied {
		return "", ErrCameraStarting
	}
	path := domain.PhotoPath(project, alias)
	return path, SaveFramePNG(frame, path) // encoded outside the view's lock, so frames keep coming
}

// Validate checks that alias is a valid photo name not used yet in the project.
func (service *CaptureService) Validate(project, alias string) error {
	if err := domain.ValidatePhotoAlias(alias); err != nil {
		return err
	}
	if _, err := os.Stat(domain.PhotoPath(project, alias)); err == nil {
		return fmt.Errorf("%w: %s", domain.ErrAliasTaken, alias)
	}
	return nil
}

// Capture saves a photo named alias in the project, streaming the capture's messages into a Run.
func (service *CaptureService) Capture(project, alias string) (*Run, error) {
	if err := service.Validate(project, alias); err != nil {
		return nil, err
	}
	if err := capture.WriteScripts(service.scriptDir); err != nil {
		return nil, err
	}
	command := Command{
		Name: "guardar imagen", Runner: "python", Program: filepath.Join(project, domain.VenvPython),
		Args: []string{filepath.Join(service.scriptDir, capture.CaptureScript), domain.PhotoPath(project, alias)}, Dir: project,
	}
	return service.runs.StartCommand(captureRunID(project), command)
}

// LatestCapture returns the last capture started for the project in this session.
func (service *CaptureService) LatestCapture(project string) (*Run, bool) {
	return service.runs.Latest(captureRunID(project))
}

func captureRunID(project string) string { return "capture:" + filepath.Clean(project) }

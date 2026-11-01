package services

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/capture"
)

// ErrCameraBusy is returned when the live view is asked to start for one project while it shows another.
var ErrCameraBusy = errors.New("the camera is showing another project")

// previewLogLimit is how many message lines the live view keeps.
const previewLogLimit = 200

// PreviewService runs one live camera view at a time and keeps its latest frame and messages.
type PreviewService struct {
	scriptDir string
	onChange  func()
	mutex     sync.Mutex
	project   string
	command   *exec.Cmd
	running   bool
	stopping  bool
	latest    domain.Frame
	sequence  uint64
	log       *domain.OutputBuffer
}

// NewPreviewService builds a service that writes its scripts into scriptDir and calls onChange on every frame and message.
func NewPreviewService(scriptDir string, onChange func()) *PreviewService {
	return &PreviewService{scriptDir: scriptDir, onChange: onChange, log: domain.NewOutputBuffer(previewLogLimit)}
}

// Start opens the live view for the project with its own Python; starting it again for the same project does nothing.
func (service *PreviewService) Start(project string) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.running {
		if service.project == project {
			return nil
		}
		return ErrCameraBusy
	}
	if err := capture.WriteScripts(service.scriptDir); err != nil {
		return err
	}
	command := exec.Command(filepath.Join(project, domain.VenvPython), filepath.Join(service.scriptDir, capture.PreviewScript))
	command.Dir = project
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("opening the camera's frames: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return fmt.Errorf("opening the camera's messages: %w", err)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("starting the camera view: %w", err)
	}
	service.project, service.command, service.running, service.stopping = project, command, true, false
	service.latest, service.log = domain.Frame{}, domain.NewOutputBuffer(previewLogLimit)
	go service.watch(stdout, stderr)
	service.onChange()
	return nil
}

// Stop closes the live view, releasing the camera; stopping a view that is not running does nothing.
func (service *PreviewService) Stop() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if !service.running || service.stopping {
		return nil
	}
	service.stopping = true
	pid := service.command.Process.Pid
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("stopping the camera view %d: %w", pid, err)
	}
	return nil
}

// Running tells whether the live view is open for this project.
func (service *PreviewService) Running(project string) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	return service.running && service.project == project
}

// Log returns the live view's messages for this project, oldest first.
func (service *PreviewService) Log(project string) []domain.Line {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.project != project {
		return nil
	}
	return service.log.Lines()
}

// WithLatest calls use with the latest frame of the project's live view and its sequence number,
// which grows with each new frame; the frame is only valid inside use. It returns false when there is none.
func (service *PreviewService) WithLatest(project string, use func(frame domain.Frame, sequence uint64)) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if !service.running || service.project != project || service.latest.RGB == nil {
		return false
	}
	use(service.latest, service.sequence)
	return true
}

package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// ErrNoPython is returned when the Python used to create environments is not installed.
var ErrNoPython = errors.New("python not found")

// checkTimeout bounds the import check, in case the module does slow work when imported.
const checkTimeout = 20 * time.Second

// PythonEnvService prepares and checks the .venv of a Python project whose module is run with `python -m`.
type PythonEnvService struct {
	runs   *RunService
	python string
	module string
}

// NewPythonEnvService builds a service that creates environments with python and checks them by importing module.
func NewPythonEnvService(runs *RunService, python, module string) *PythonEnvService {
	return &PythonEnvService{runs: runs, python: python, module: module}
}

// Module returns the name of the module the environment must be able to import.
func (service *PythonEnvService) Module() string { return service.module }

// Check says how ready the project in dir is and, when the module does not import, why; it runs the
// environment's Python once, so call it on events, not every frame.
func (service *PythonEnvService) Check(dir string) (domain.EnvCheck, error) {
	if dir == "" || !isDir(dir) {
		return domain.EnvCheck{Status: domain.NoProject}, nil
	}
	if _, err := os.Stat(filepath.Join(dir, domain.VenvPython)); err != nil {
		return domain.EnvCheck{Status: domain.NoVenv}, nil
	}
	deadline, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	command := exec.CommandContext(deadline, domain.VenvPython, "-c", "import "+service.module)
	command.Dir = dir
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err := command.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return domain.EnvCheck{Status: domain.Ready}, nil
	case errors.As(err, &exitErr):
		return domain.EnvCheck{Status: domain.Incomplete, Reason: domain.DescribeImportFailure(service.module, stderr.String())}, nil
	default:
		return domain.EnvCheck{Status: domain.Incomplete}, fmt.Errorf("checking the environment in %s: %w", dir, err)
	}
}

// Setup creates the environment in dir and installs its requirements, streaming the output into a Run.
func (service *PythonEnvService) Setup(dir string) (*Run, error) {
	if !isDir(dir) {
		return nil, fmt.Errorf("cannot prepare %s: %w", dir, os.ErrNotExist)
	}
	if _, err := exec.LookPath(service.python); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNoPython, service.python)
	}
	_, err := os.Stat(filepath.Join(dir, domain.RequirementsFile))
	script := domain.SetupScript(domain.SetupPlan(service.python, err == nil))
	return service.runs.StartCommand(setupRunID(dir), Command{Name: "preparar entorno", Runner: "bash", Program: "bash", Args: []string{"-c", script}, Dir: dir})
}

// LatestSetup returns the last setup run started for dir in this session.
func (service *PythonEnvService) LatestSetup(dir string) (*Run, bool) {
	return service.runs.Latest(setupRunID(dir))
}

func setupRunID(dir string) string { return "setup:" + filepath.Clean(dir) }

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

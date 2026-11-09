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

// ErrNoPython is returned when the Python used to create environments is not installed and none can be downloaded.
var ErrNoPython = errors.New("python not found")

// checkTimeout bounds the import check, in case the module does slow work when imported.
const checkTimeout = 20 * time.Second

// PythonEnvService prepares and checks the .venv of a Python project whose module is run with `python -m`.
type PythonEnvService struct {
	runs        *RunService
	python      string
	module      string
	download    *domain.PythonDownload // fetched into downloadDir when python is not installed
	downloadDir string
}

// NewPythonEnvService builds a service that creates environments with python and checks them by importing module.
func NewPythonEnvService(runs *RunService, python, module string) *PythonEnvService {
	return &PythonEnvService{runs: runs, python: python, module: module}
}

// WithDownload lets Setup fetch download into dir when the computer has no Python, and use it from
// then on.
func (service *PythonEnvService) WithDownload(download domain.PythonDownload, dir string) *PythonEnvService {
	service.download, service.downloadDir = &download, dir
	return service
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
		_, missing := os.Stat(filepath.Join(dir, domain.ProjectMarker))
		reason := domain.DescribeImportFailure(service.module, stderr.String(), missing == nil)
		return domain.EnvCheck{Status: domain.Incomplete, Reason: reason}, nil
	default:
		return domain.EnvCheck{Status: domain.Incomplete}, fmt.Errorf("checking the environment in %s: %w", dir, err)
	}
}

// Setup creates the environment in dir and installs its requirements, streaming the output into a Run.
func (service *PythonEnvService) Setup(dir string) (*Run, error) {
	if !isDir(dir) {
		return nil, fmt.Errorf("cannot prepare %s: %w", dir, os.ErrNotExist)
	}
	python, install, err := service.pickPython()
	if err != nil {
		return nil, err
	}
	requirements, _ := os.ReadFile(filepath.Join(dir, domain.RequirementsFile)) // missing reads as empty
	script := domain.SetupScript(append(install, domain.SetupPlan(python, domain.ListsPackages(string(requirements)))...))
	return service.runs.StartCommand(setupRunID(dir), Command{Name: "preparar entorno", Runner: "bash", Program: "bash", Args: []string{"-c", script}, Dir: dir})
}

// pickPython chooses the Python to create the environment with: one downloaded earlier, so every
// project keeps the same one, else the computer's own, else a fresh download, whose steps it returns.
func (service *PythonEnvService) pickPython() (string, []domain.SetupStep, error) {
	if service.download != nil {
		if downloaded := domain.DownloadedPython(service.downloadDir); isFile(downloaded) {
			return downloaded, nil, nil
		}
	}
	if _, err := exec.LookPath(service.python); err == nil {
		return service.python, nil, nil
	}
	if service.download == nil {
		return "", nil, fmt.Errorf("%w: %s", ErrNoPython, service.python)
	}
	return domain.DownloadedPython(service.downloadDir), domain.PythonInstallSteps(*service.download, service.downloadDir), nil
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

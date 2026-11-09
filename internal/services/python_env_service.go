package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	platform    domain.Platform        // picks which requirements file is installed
	download    *domain.PythonDownload // fetched into downloadDir when python is not installed
	downloadDir string
}

// NewPythonEnvService builds a service that creates environments with python and checks them by importing module.
func NewPythonEnvService(runs *RunService, python, module string) *PythonEnvService {
	return &PythonEnvService{runs: runs, python: python, module: module, platform: domain.PlatformFor(runtime.GOOS)}
}

// Platform returns the computer family whose requirements file this service installs.
func (service *PythonEnvService) Platform() domain.Platform { return service.platform }

// Requirements returns the requirements file Setup installs from dir: this computer's own, such as
// requirements-macos.txt, when the project has it, requirements.txt otherwise, and "" when neither
// exists. It looks at the disk, so call it on events, not every frame.
func (service *PythonEnvService) Requirements(dir string) string {
	return domain.RequirementsFor(service.platform, func(name string) bool { return isFile(filepath.Join(dir, name)) })
}

// WithDownload makes the downloaded Python the one environments are created with: InstallPython
// fetches it into dir, and Setup fetches it first when it is not there yet. The computer's own
// Python is then not used, so every computer runs the same, tested version.
func (service *PythonEnvService) WithDownload(download domain.PythonDownload, dir string) *PythonEnvService {
	service.download, service.downloadDir = &download, dir
	return service
}

// Download returns the Python this service installs, if it has one.
func (service *PythonEnvService) Download() (domain.PythonDownload, bool) {
	if service.download == nil {
		return domain.PythonDownload{}, false
	}
	return *service.download, true
}

// PythonInstalled tells whether the downloaded Python is in place; it looks at the disk, so call it
// on events, not every frame.
func (service *PythonEnvService) PythonInstalled() bool {
	return service.download != nil && isFile(domain.DownloadedPython(service.downloadDir))
}

// InstallPython downloads the Python on its own, streaming the output into a Run.
func (service *PythonEnvService) InstallPython() (*Run, error) {
	if service.download == nil {
		return nil, fmt.Errorf("%w: no download for this computer", ErrNoPython)
	}
	parent := filepath.Dir(service.downloadDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("making room for Python in %s: %w", parent, err)
	}
	script := domain.InstallScript(domain.PythonInstallSteps(*service.download, service.downloadDir)) + "echo '✓ Python instalado'\n"
	return service.runs.StartCommand(pythonInstallRunID, Command{Name: "instalar Python", Runner: "bash", Program: "bash", Args: []string{"-c", script}, Dir: parent})
}

// LatestPythonInstall returns the last Python installation started in this session.
func (service *PythonEnvService) LatestPythonInstall() (*Run, bool) {
	return service.runs.Latest(pythonInstallRunID)
}

// pythonInstallRunID names the run that installs Python; there is one Python for every project.
const pythonInstallRunID = "install-python"

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
	if run, ok := service.LatestPythonInstall(); ok && run.Running() {
		return nil, fmt.Errorf("%w: instalar Python", ErrAlreadyRunning)
	}
	python, install, err := service.pickPython()
	if err != nil {
		return nil, err
	}
	requirements := service.Requirements(dir)
	content, _ := os.ReadFile(filepath.Join(dir, requirements)) // none reads as empty
	script := domain.SetupScript(append(install, domain.SetupPlan(python, requirements, domain.ListsPackages(string(content)))...))
	return service.runs.StartCommand(setupRunID(dir), Command{Name: "preparar entorno", Runner: "bash", Program: "bash", Args: []string{"-c", script}, Dir: dir})
}

// pickPython chooses the Python to create the environment with: the downloaded one, with the steps
// that fetch it when it is not there yet, or, on a computer with no download, its own.
func (service *PythonEnvService) pickPython() (string, []domain.SetupStep, error) {
	if service.download != nil {
		downloaded := domain.DownloadedPython(service.downloadDir)
		if isFile(downloaded) {
			return downloaded, nil, nil
		}
		return downloaded, domain.PythonInstallSteps(*service.download, service.downloadDir), nil
	}
	if _, err := exec.LookPath(service.python); err != nil {
		return "", nil, fmt.Errorf("%w: %s", ErrNoPython, service.python)
	}
	return service.python, nil, nil
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

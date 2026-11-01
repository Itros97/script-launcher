package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// OutputLimit is how many lines of output each run keeps.
const OutputLimit = 2000

// RunService starts and stops scripts and keeps the latest run of each one.
type RunService struct {
	registry domain.Registry
	now      func() time.Time
	onChange func()
	mutex    sync.Mutex
	runs     map[string]*Run
}

// NewRunService builds a service that picks runners from registry and calls onChange whenever a run changes (from any goroutine).
func NewRunService(registry domain.Registry, now func() time.Time, onChange func()) *RunService {
	return &RunService{registry: registry, now: now, onChange: onChange, runs: map[string]*Run{}}
}

// Resolve checks that path is a regular file a runner can launch, and returns that runner.
func (service *RunService) Resolve(path string) (domain.Runner, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot use %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", ErrNotAFile, path)
	}
	return service.registry.Resolve(path, info.Mode().Perm()&0o111 != 0)
}

// Start launches a saved script in its own folder with the runner that matches it.
func (service *RunService) Start(script domain.Script) (*Run, error) {
	runner, err := service.Resolve(script.Path)
	if err != nil {
		return nil, err
	}
	program, args := runner.Command(script.Path)
	return service.StartCommand(script.ID, Command{Name: script.Name, Runner: runner.ID(), Program: program, Args: args, Dir: filepath.Dir(script.Path)})
}

// StartCommand launches a command in its own process group, streaming its output into a new Run kept under id.
func (service *RunService) StartCommand(id string, spec Command) (*Run, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if last, ok := service.runs[id]; ok && last.Running() {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyRunning, spec.Name)
	}
	command := exec.Command(spec.Program, spec.Args...)
	command.Dir = spec.Dir
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening output of %s: %w", spec.Name, err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("opening errors of %s: %w", spec.Name, err)
	}
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("starting %s with %s: %w", spec.Name, spec.Runner, err)
	}
	run := &Run{
		command: command, output: domain.NewOutputBuffer(OutputLimit), status: Running, runner: spec.Runner,
		startedAt: service.now(), now: service.now, onChange: service.onChange,
	}
	service.runs[id] = run
	go service.watch(run, stdout, stderr)
	service.onChange()
	return run, nil
}

// watch drains both pipes, then waits for the process so its exit is recorded after its last line.
func (service *RunService) watch(run *Run, stdout, stderr io.Reader) {
	var group sync.WaitGroup
	group.Add(2)
	go func() { defer group.Done(); run.read(stdout, domain.Stdout) }()
	go func() { defer group.Done(); run.read(stderr, domain.Stderr) }()
	group.Wait()
	run.finish(run.command.Wait())
}

// Stop sends SIGTERM to the script's whole process group; stopping a script that is not running does nothing.
func (service *RunService) Stop(scriptID string) error {
	run, ok := service.Latest(scriptID)
	if !ok || !run.Running() {
		return nil
	}
	run.mutex.Lock()
	run.stopping = true
	pid := run.command.Process.Pid
	run.mutex.Unlock()
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("stopping process group %d: %w", pid, err)
	}
	return nil
}

// Latest returns the most recent run of a script, if it was ever started.
func (service *RunService) Latest(scriptID string) (*Run, bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	run, ok := service.runs[scriptID]
	return run, ok
}

// StopAll stops every running script, for when the app closes, and reports the ones that could not be stopped.
func (service *RunService) StopAll() error {
	service.mutex.Lock()
	ids := make([]string, 0, len(service.runs))
	for id := range service.runs {
		ids = append(ids, id)
	}
	service.mutex.Unlock()
	var failures []error
	for _, id := range ids {
		failures = append(failures, service.Stop(id))
	}
	return errors.Join(failures...)
}

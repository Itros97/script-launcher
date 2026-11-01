package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// RunStatus is where a run is in its life.
type RunStatus int

const (
	Running   RunStatus = iota // the process is alive
	Succeeded                  // it exited with code 0
	Failed                     // it exited with another code, or could not be waited on
	Stopped                    // the user stopped it
)

// RunSnapshot is a copy of a run's state, safe to read while the process keeps writing.
type RunSnapshot struct {
	Status    RunStatus
	Runner    string
	ExitCode  int
	StartedAt time.Time
	Duration  time.Duration
	Lines     []domain.Line
	Dropped   int
}

// Run is one execution of a script; its output and status are filled in from the process's goroutines.
type Run struct {
	mutex     sync.Mutex
	command   *exec.Cmd
	output    *domain.OutputBuffer
	status    RunStatus
	runner    string
	exitCode  int
	startedAt time.Time
	endedAt   time.Time
	stopping  bool
	now       func() time.Time
	onChange  func()
}

// Snapshot copies the run's current state for the UI.
func (run *Run) Snapshot() RunSnapshot {
	run.mutex.Lock()
	defer run.mutex.Unlock()
	end := run.endedAt
	if run.status == Running {
		end = run.now()
	}
	return RunSnapshot{
		Status: run.status, Runner: run.runner, ExitCode: run.exitCode, StartedAt: run.startedAt,
		Duration: end.Sub(run.startedAt), Lines: run.output.Lines(), Dropped: run.output.Dropped(),
	}
}

// Running tells whether the process is still alive.
func (run *Run) Running() bool {
	run.mutex.Lock()
	defer run.mutex.Unlock()
	return run.status == Running
}

func (run *Run) append(line domain.Line) {
	run.mutex.Lock()
	run.output.Append(line)
	run.mutex.Unlock()
	run.onChange()
}

// read copies one of the process's pipes into the output, line by line, until it closes.
func (run *Run) read(pipe io.Reader, stream domain.Stream) {
	reader := bufio.NewReader(pipe)
	for {
		text, err := reader.ReadString('\n')
		if text != "" {
			run.append(domain.Line{Text: cleanLine(text), Stream: stream})
		}
		if err != nil {
			return
		}
	}
}

// finish records how the process ended once both pipes are drained.
func (run *Run) finish(waitErr error) {
	run.mutex.Lock()
	run.endedAt = run.now()
	run.exitCode = run.command.ProcessState.ExitCode()
	var exitErr *exec.ExitError
	switch {
	case run.stopping:
		run.status = Stopped
	case waitErr == nil:
		run.status = Succeeded
	case errors.As(waitErr, &exitErr):
		run.status = Failed
	default:
		run.status = Failed
		run.output.Append(domain.Line{Text: fmt.Sprintf("waiting for the process failed: %v", waitErr), Stream: domain.System})
	}
	run.output.Append(domain.Line{Text: run.summary(), Stream: domain.System})
	run.mutex.Unlock()
	run.onChange()
}

func (run *Run) summary() string {
	duration := run.endedAt.Sub(run.startedAt).Round(time.Millisecond)
	switch run.status {
	case Stopped:
		return fmt.Sprintf("■ stopped after %s", duration)
	case Succeeded:
		return fmt.Sprintf("✓ exited with code 0 in %s", duration)
	default:
		return fmt.Sprintf("✗ exited with code %d in %s", run.exitCode, duration)
	}
}

// cleanLine drops the line ending and keeps only what a terminal would show after a carriage return.
func cleanLine(text string) string {
	text = strings.TrimRight(text, "\r\n")
	if index := strings.LastIndex(text, "\r"); index >= 0 && index < len(text)-1 {
		text = text[index+1:]
	}
	return strings.ReplaceAll(text, "\t", "    ")
}

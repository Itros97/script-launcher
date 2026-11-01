package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// ErrUnsupportedScript is returned when no runner knows how to launch a file.
var ErrUnsupportedScript = errors.New("unsupported script")

// Runner knows how to turn a script path into the command that launches it.
type Runner interface {
	// ID names the runner, shown in the UI ("bash", "python3", "executable").
	ID() string
	// Matches tells whether this runner can launch the file.
	Matches(path string, executable bool) bool
	// Command returns the program and arguments that run the file.
	Command(path string) (program string, args []string)
}

// InterpreterRunner launches files with a given extension through an interpreter program.
type InterpreterRunner struct {
	Program    string
	Extensions []string
}

// ID returns the interpreter program name.
func (runner InterpreterRunner) ID() string { return runner.Program }

// Matches accepts files whose extension is in the runner's list, case-insensitively.
func (runner InterpreterRunner) Matches(path string, _ bool) bool {
	return slices.Contains(runner.Extensions, strings.ToLower(filepath.Ext(path)))
}

// Command runs the file as the interpreter's first argument.
func (runner InterpreterRunner) Command(path string) (string, []string) {
	return runner.Program, []string{path}
}

// ExecutableRunner launches any file with the executable bit set, directly.
type ExecutableRunner struct{}

// ID returns "executable".
func (ExecutableRunner) ID() string { return "executable" }

// Matches accepts any executable file.
func (ExecutableRunner) Matches(_ string, executable bool) bool { return executable }

// Command runs the file itself with no arguments.
func (ExecutableRunner) Command(path string) (string, []string) { return path, nil }

// Registry picks the first runner that matches a file, in the order the runners were given.
type Registry struct {
	runners []Runner
}

// NewRegistry builds a registry that tries runners in order.
func NewRegistry(runners ...Runner) Registry { return Registry{runners: runners} }

// DefaultRegistry knows bash, python3 and node by extension, and falls back to executables.
func DefaultRegistry() Registry {
	return NewRegistry(
		InterpreterRunner{Program: "bash", Extensions: []string{".sh", ".bash"}},
		InterpreterRunner{Program: "python3", Extensions: []string{".py"}},
		InterpreterRunner{Program: "node", Extensions: []string{".js", ".mjs", ".ts"}},
		ExecutableRunner{},
	)
}

// Resolve returns the runner for a file, or ErrUnsupportedScript when none matches.
func (registry Registry) Resolve(path string, executable bool) (Runner, error) {
	for _, runner := range registry.runners {
		if runner.Matches(path, executable) {
			return runner, nil
		}
	}
	return nil, fmt.Errorf("%w: %s has no known extension and is not executable", ErrUnsupportedScript, filepath.Base(path))
}

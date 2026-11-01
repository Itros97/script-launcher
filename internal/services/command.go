package services

import "errors"

var (
	// ErrAlreadyRunning is returned when a command is started while its last run is still alive.
	ErrAlreadyRunning = errors.New("already running")
	// ErrNotAFile is returned when a script path points to something that is not a regular file.
	ErrNotAFile = errors.New("not a regular file")
	// ErrTemplateExists is returned when both names a template can be saved under are taken.
	ErrTemplateExists = errors.New("template already saved")
)

// Command is a program to run in a folder; Name and Runner describe it in messages and in the UI.
type Command struct {
	Name    string
	Runner  string
	Program string
	Args    []string
	Dir     string
}

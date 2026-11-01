package views

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// shortPath writes the home folder as ~, as a shell prompt does.
func shortPath(path, home string) string {
	if home == "" || (path != home && !strings.HasPrefix(path, home+string(filepath.Separator))) {
		return path
	}
	return "~" + strings.TrimPrefix(path, home)
}

// formatDuration shows short runs in milliseconds and longer ones to a tenth of a second.
func formatDuration(duration time.Duration) string {
	if duration < time.Second {
		return duration.Round(time.Millisecond).String()
	}
	return duration.Round(100 * time.Millisecond).String()
}

// fileExists tells whether a saved script is still on disk.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

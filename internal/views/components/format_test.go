package components

import (
	"testing"
	"time"
)

func TestShortPathWritesHomeAsTilde(t *testing.T) {
	cases := map[string]string{"/home/user/bin": "~/bin", "/home/user": "~", "/home/username": "/home/username", "/opt/x": "/opt/x"}
	for path, want := range cases {
		if got := ShortPath(path, "/home/user"); got != want {
			t.Errorf("ShortPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestFormatDurationKeepsShortRunsReadable(t *testing.T) {
	cases := map[time.Duration]string{42 * time.Millisecond: "42ms", 1540 * time.Millisecond: "1.5s", 90 * time.Second: "1m30s"}
	for duration, want := range cases {
		if got := formatDuration(duration); got != want {
			t.Errorf("formatDuration(%s) = %q, want %q", duration, got, want)
		}
	}
}

package services

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

func newService() *RunService {
	return NewRunService(domain.DefaultRegistry(), time.Now, func() {})
}

func fixture(t *testing.T, name string) domain.Script {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return domain.NewScript(path, time.Now())
}

// waitFor polls the run until it stops running, failing after a timeout.
func waitFor(t *testing.T, run *Run) RunSnapshot {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for run.Running() {
		if time.Now().After(deadline) {
			t.Fatal("run did not finish in time")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return run.Snapshot()
}

func texts(lines []domain.Line, stream domain.Stream) []string {
	var found []string
	for _, line := range lines {
		if line.Stream == stream {
			found = append(found, line.Text)
		}
	}
	return found
}

func TestRunCapturesOutputErrorsAndExitCode(t *testing.T) {
	run, err := newService().Start(fixture(t, "fails.sh"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Failed || snapshot.ExitCode != 3 || snapshot.Runner != "bash" {
		t.Errorf("got status %d, code %d, runner %s", snapshot.Status, snapshot.ExitCode, snapshot.Runner)
	}
	if got := texts(snapshot.Lines, domain.Stdout); len(got) != 1 || got[0] != "hello" {
		t.Errorf("stdout: got %v", got)
	}
	if got := texts(snapshot.Lines, domain.Stderr); len(got) != 1 || got[0] != "oops" {
		t.Errorf("stderr: got %v", got)
	}
}

func TestRunsExecutableWithoutExtensionDirectly(t *testing.T) {
	run, err := newService().Start(fixture(t, "tool"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Succeeded || snapshot.Runner != "executable" || texts(snapshot.Lines, domain.Stdout)[0] != "direct" {
		t.Errorf("got %+v", snapshot)
	}
}

func TestStopEndsTheRunAndAllowsANewStart(t *testing.T) {
	service := newService()
	script := fixture(t, "sleeps.sh")
	run, err := service.Start(script)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Start(script); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("second start: got %v, want ErrAlreadyRunning", err)
	}
	if err := service.Stop(script.ID); err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Stopped || snapshot.Duration > 5*time.Second {
		t.Errorf("got status %d after %s", snapshot.Status, snapshot.Duration)
	}
	if strings.Contains(strings.Join(texts(snapshot.Lines, domain.Stdout), "\n"), "never") {
		t.Error("script kept running after stop")
	}
	if _, err := service.Start(script); err != nil {
		t.Errorf("start after stop: %v", err)
	}
	_ = service.StopAll()
}

func TestResolveRejectsMissingAndUnsupportedFiles(t *testing.T) {
	service := newService()
	if _, err := service.Resolve(filepath.Join("testdata", "missing.sh")); err == nil {
		t.Error("expected an error for a missing file")
	}
	if _, err := service.Resolve(filepath.Join("testdata", "notes.txt")); !errors.Is(err, domain.ErrUnsupportedScript) {
		t.Errorf("got %v, want ErrUnsupportedScript", err)
	}
	if _, err := service.Resolve("testdata"); err == nil {
		t.Error("expected an error for a folder")
	}
}

func TestCleanLineKeepsWhatATerminalShows(t *testing.T) {
	cases := map[string]string{"plain\n": "plain", "dos\r\n": "dos", "10%\r50%\r100%\n": "100%", "a\tb": "a    b"}
	for input, want := range cases {
		if got := cleanLine(input); got != want {
			t.Errorf("cleanLine(%q) = %q, want %q", input, got, want)
		}
	}
}

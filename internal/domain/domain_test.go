package domain

import (
	"errors"
	"testing"
	"time"
)

func TestResolvesInterpreterByExtension(t *testing.T) {
	registry := DefaultRegistry()
	cases := map[string]string{
		"/scripts/backup.sh": "bash",
		"/scripts/report.PY": "python3",
		"/scripts/tool.js":   "node",
		"/scripts/tool.ts":   "node",
	}
	for path, want := range cases {
		runner, err := registry.Resolve(path, false)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", path, err)
		}
		if runner.ID() != want {
			t.Errorf("%s: got runner %s, want %s", path, runner.ID(), want)
		}
	}
}

func TestPrefersInterpreterOverExecutableBit(t *testing.T) {
	runner, err := DefaultRegistry().Resolve("/scripts/run.sh", true)
	if err != nil || runner.ID() != "bash" {
		t.Fatalf("got %v, %v; want bash", runner, err)
	}
}

func TestFallsBackToExecutableWhenExtensionIsUnknown(t *testing.T) {
	runner, err := DefaultRegistry().Resolve("/scripts/deploy", true)
	if err != nil {
		t.Fatal(err)
	}
	program, args := runner.Command("/scripts/deploy")
	if runner.ID() != "executable" || program != "/scripts/deploy" || len(args) != 0 {
		t.Errorf("got %s %s %v", runner.ID(), program, args)
	}
}

func TestRejectsUnknownNonExecutableFile(t *testing.T) {
	_, err := DefaultRegistry().Resolve("/scripts/notes.txt", false)
	if !errors.Is(err, ErrUnsupportedScript) {
		t.Fatalf("got %v, want ErrUnsupportedScript", err)
	}
}

func TestInterpreterCommandPassesPathAsArgument(t *testing.T) {
	program, args := InterpreterRunner{Program: "bash"}.Command("/scripts/a.sh")
	if program != "bash" || len(args) != 1 || args[0] != "/scripts/a.sh" {
		t.Errorf("got %s %v", program, args)
	}
}

func TestSamePathGivesSameScriptID(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first := NewScript("/scripts/../scripts/a.sh", at)
	second := NewScript("/scripts/a.sh", at)
	if first.ID != second.ID || first.Path != "/scripts/a.sh" || first.Name != "a.sh" {
		t.Errorf("got %+v and %+v", first, second)
	}
}

func TestOutputBufferDropsOldestLinesPastLimit(t *testing.T) {
	buffer := NewOutputBuffer(2)
	for _, text := range []string{"one", "two", "three"} {
		buffer.Append(Line{Text: text})
	}
	lines := buffer.Lines()
	if len(lines) != 2 || lines[0].Text != "two" || lines[1].Text != "three" || buffer.Dropped() != 1 {
		t.Errorf("got %+v, dropped %d", lines, buffer.Dropped())
	}
}

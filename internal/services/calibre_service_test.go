package services

import (
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

// copyTree copies a folder of small files.
func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(from, path)
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(to, relative), 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, relative), content, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// fakeCalibreProject creates a project with a .venv and the fake calibre3d package.
func fakeCalibreProject(t *testing.T) string {
	t.Helper()
	project := t.TempDir()
	copyTree(t, filepath.Join("testdata", "fake_calibre3d"), project)
	if output, err := exec.Command("python3", "-m", "venv", filepath.Join(project, domain.VenvDir)).CombinedOutput(); err != nil {
		t.Fatalf("creating venv: %v\n%s", err, output)
	}
	return project
}

func newCalibreService(t *testing.T) *CalibreService {
	return NewCalibreService(newService(), NewPreviewService(t.TempDir(), func() {}))
}

// runStep runs a step to its end and returns how it ended.
func runStep(t *testing.T, service *CalibreService, project string, step domain.CalibreStep, name string) RunSnapshot {
	t.Helper()
	run, _, err := service.Run(project, step, name)
	if err != nil {
		t.Fatal(err)
	}
	return waitFor(t, run)
}

func TestMeasuringLeavesAResultTheLauncherCanRead(t *testing.T) {
	project := fakeCalibreProject(t)
	service := newCalibreService(t)
	if snapshot := runStep(t, service, project, domain.StepMeasure, "pieza1"); snapshot.Status != Succeeded {
		t.Fatalf("measuring failed: %+v", snapshot.Lines)
	}
	measurement, err := service.LoadMeasurement(project, domain.StepMeasure, "pieza1")
	if err != nil || measurement.Length != 61.234 || len(measurement.Warnings) != 1 {
		t.Errorf("got %+v, %v", measurement, err)
	}
	if _, err := os.Stat(domain.StepMeasure.ResultFile(project, "pieza1", "_mapa")); err != nil {
		t.Errorf("no heat map: %v", err)
	}
	if _, _, err := service.Run(project, domain.StepMeasure, "pieza1"); !errors.Is(err, domain.ErrAliasTaken) {
		t.Errorf("measuring the same name again: got %v, want ErrAliasTaken", err)
	}
}

func TestThePlaneIsReadOnceSaved(t *testing.T) {
	project := fakeCalibreProject(t)
	service := newCalibreService(t)
	if _, ok, err := service.LoadPlane(project); ok || err != nil {
		t.Fatalf("before the plane: got ok=%v, %v", ok, err)
	}
	runStep(t, service, project, domain.StepPlane, "")
	plane, ok, err := service.LoadPlane(project)
	if !ok || err != nil || plane.NoiseMM != 0.04 {
		t.Errorf("got %+v, %v, %v", plane, ok, err)
	}
}

func TestTheDemoLeavesItsResultInItsOwnFolder(t *testing.T) {
	project := fakeCalibreProject(t)
	service := newCalibreService(t)
	runStep(t, service, project, domain.StepDemo, "")
	if _, err := service.LoadMeasurement(project, domain.StepDemo, ""); err != nil {
		t.Error(err)
	}
}

func TestAFailingStepShowsWhy(t *testing.T) {
	project := fakeCalibreProject(t)
	t.Setenv("FAKE_CALIBRE_FAIL", "1")
	snapshot := runStep(t, newCalibreService(t), project, domain.StepPlane, "")
	if snapshot.Status != Failed || !strings.Contains(joinLines(snapshot.Lines), "No device connected") {
		t.Errorf("got %d: %+v", snapshot.Status, snapshot.Lines)
	}
}

func TestAStepClosesTheLiveCameraFirst(t *testing.T) {
	project := projectWithFakeCamera(t)
	copyTree(t, filepath.Join("testdata", "fake_calibre3d"), project)
	preview := NewPreviewService(t.TempDir(), func() {})
	service := NewCalibreService(newService(), preview)
	if err := preview.Start(project); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, "a first frame", func() bool { return preview.WithLatest(project, func(domain.Frame, uint64) {}) })
	run, closed, err := service.Run(project, domain.StepCenter, "")
	if err != nil || !closed || preview.Running(project) {
		t.Fatalf("got closed=%v running=%v, %v", closed, preview.Running(project), err)
	}
	waitFor(t, run)
}

// TestTheRealCalibre3DDemo runs the real project's demo when CALIBRE3D_PROJECT names its folder (with its .venv).
func TestTheRealCalibre3DDemo(t *testing.T) {
	project := os.Getenv("CALIBRE3D_PROJECT")
	if project == "" {
		t.Skip("set CALIBRE3D_PROJECT to the calibre3d project's folder to run it")
	}
	service := newCalibreService(t)
	if snapshot := runStep(t, service, project, domain.StepDemo, ""); snapshot.Status != Succeeded {
		t.Fatalf("demo failed: %s", joinLines(snapshot.Lines))
	}
	measurement, err := service.LoadMeasurement(project, domain.StepDemo, "")
	if err != nil || math.Abs(measurement.Length-60) > 0.5 || math.Abs(measurement.Width-40) > 0.5 || math.Abs(measurement.Thickness-20) > 0.5 {
		t.Errorf("got %+v, %v", measurement, err)
	}
}

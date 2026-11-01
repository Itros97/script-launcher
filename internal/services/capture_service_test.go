package services

import (
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

// projectWithFakeCamera creates a project with a real .venv whose pyrealsense2 is the fake in testdata.
func projectWithFakeCamera(t *testing.T) string {
	t.Helper()
	dir := fakeProject(t)
	run, err := newEnvService("python3").Setup(dir)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := waitFor(t, run); snapshot.Status != Succeeded {
		t.Fatalf("venv setup failed: %+v", snapshot.Lines)
	}
	sitePackages, err := filepath.Glob(filepath.Join(dir, ".venv", "lib", "python3*", "site-packages"))
	if err != nil || len(sitePackages) != 1 {
		t.Fatalf("site-packages not found: %v %v", sitePackages, err)
	}
	fake, err := os.ReadFile(filepath.Join("testdata", "fake_realsense", "pyrealsense2.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sitePackages[0], "pyrealsense2.py"), fake, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newCaptureService(t *testing.T) *CaptureService {
	scripts := t.TempDir()
	return NewCaptureService(newService(), NewPreviewService(scripts, func() {}), scripts)
}

func TestCaptureSavesA1280By720PNG(t *testing.T) {
	project := projectWithFakeCamera(t)
	service := newCaptureService(t)
	run, err := service.Capture(project, "pieza1")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := waitFor(t, run); snapshot.Status != Succeeded {
		t.Fatalf("capture failed: %+v", snapshot.Lines)
	}
	file, err := os.Open(domain.PhotoPath(project, "pieza1"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	picture, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	red, green, blue, _ := picture.At(640, 360).RGBA()
	if bounds := picture.Bounds(); bounds.Dx() != 1280 || bounds.Dy() != 720 || red>>8 != 255 || green>>8 != 128 || blue>>8 != 0 {
		t.Errorf("got %v with colour %d %d %d", bounds, red>>8, green>>8, blue>>8)
	}
	if _, err := service.Capture(project, "pieza1"); !errors.Is(err, domain.ErrAliasTaken) {
		t.Errorf("second capture with the same name: got %v, want ErrAliasTaken", err)
	}
}

func TestCaptureSaysWhenNoCameraIsConnected(t *testing.T) {
	project := projectWithFakeCamera(t)
	t.Setenv("FAKE_RS_DEVICES", "0")
	run, err := newCaptureService(t).Capture(project, "sin_camara")
	if err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, run)
	if snapshot.Status != Failed || !strings.Contains(joinLines(snapshot.Lines), "No se encuentra ninguna RealSense") {
		t.Errorf("got %d: %+v", snapshot.Status, snapshot.Lines)
	}
	if _, err := os.Stat(domain.PhotoPath(project, "sin_camara")); err == nil {
		t.Error("a photo was saved without a camera")
	}
}

func TestCaptureSaysWhenPyrealsenseIsMissing(t *testing.T) {
	project := fakeProject(t)
	run, err := newEnvService("python3").Setup(project)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, run)
	run, err = newCaptureService(t).Capture(project, "foto")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := waitFor(t, run); !strings.Contains(joinLines(snapshot.Lines), "pyrealsense2") {
		t.Errorf("got %+v", snapshot.Lines)
	}
}

func TestCaptureRejectsAnInvalidNameBeforeRunning(t *testing.T) {
	if _, err := newCaptureService(t).Capture(fakeProject(t), "a/b"); !errors.Is(err, domain.ErrInvalidAlias) {
		t.Errorf("got %v, want ErrInvalidAlias", err)
	}
}

func joinLines(lines []domain.Line) string {
	var texts []string
	for _, line := range lines {
		texts = append(texts, line.Text)
	}
	return strings.Join(texts, "\n")
}

package services

import (
	"errors"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Itros97/script-launcher/internal/domain"
)

// waitUntil polls condition until it holds, failing after a timeout with what.
func waitUntil(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestPreviewStreamsFramesUntilStopped(t *testing.T) {
	project := projectWithFakeCamera(t)
	preview := NewPreviewService(t.TempDir(), func() {})
	if err := preview.Start(project); err != nil {
		t.Fatal(err)
	}
	var first, later uint64
	var width, height int
	var pixel []byte
	waitUntil(t, "a first frame", func() bool {
		return preview.WithLatest(project, func(frame domain.Frame, sequence uint64) {
			first, width, height, pixel = sequence, frame.Width, frame.Height, append([]byte(nil), frame.RGB[:3]...)
		})
	})
	if width != 1280 || height != 720 || pixel[0] != 255 || pixel[1] != 128 || pixel[2] != 0 {
		t.Errorf("got %d×%d with pixel %v", width, height, pixel)
	}
	waitUntil(t, "more frames", func() bool {
		preview.WithLatest(project, func(_ domain.Frame, sequence uint64) { later = sequence })
		return later > first+2
	})
	if err := preview.Stop(); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, "the view to close", func() bool { return !preview.Running(project) })
	if preview.WithLatest(project, func(domain.Frame, uint64) {}) {
		t.Error("a closed view still has a frame")
	}
	if log := joinLines(preview.Log(project)); !strings.Contains(log, "■ Cámara parada") {
		t.Errorf("got log %q", log)
	}
}

func TestPreviewSaysWhenNoCameraIsConnected(t *testing.T) {
	project := projectWithFakeCamera(t)
	t.Setenv("FAKE_RS_DEVICES", "0")
	preview := NewPreviewService(t.TempDir(), func() {})
	if err := preview.Start(project); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, "the view to close", func() bool { return !preview.Running(project) })
	log := joinLines(preview.Log(project))
	if !strings.Contains(log, "No se encuentra ninguna RealSense") || !strings.Contains(log, "✗ La cámara se ha cerrado (código 4)") {
		t.Errorf("got log %q", log)
	}
}

func TestPreviewRefusesASecondProjectWhileOpen(t *testing.T) {
	project := projectWithFakeCamera(t)
	preview := NewPreviewService(t.TempDir(), func() {})
	if err := preview.Start(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = preview.Stop() })
	if err := preview.Start(project); err != nil {
		t.Errorf("starting the same project again: %v", err)
	}
	if err := preview.Start(t.TempDir()); !errors.Is(err, ErrCameraBusy) {
		t.Errorf("got %v, want ErrCameraBusy", err)
	}
}

func TestPhotoWhileLiveIsTheFrameOnScreen(t *testing.T) {
	project := projectWithFakeCamera(t)
	scripts := t.TempDir()
	preview := NewPreviewService(scripts, func() {})
	camera := NewCaptureService(newService(), preview, scripts)
	if err := preview.Start(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = preview.Stop() })
	waitUntil(t, "a first frame", func() bool { return preview.WithLatest(project, func(domain.Frame, uint64) {}) })
	if !camera.Live(project) {
		t.Fatal("the camera should be live")
	}
	path, err := camera.SaveLive(project, "en_directo")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	picture, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	red, green, _, _ := picture.At(10, 10).RGBA()
	if picture.Bounds().Dx() != 1280 || red>>8 != 255 || green>>8 != 128 {
		t.Errorf("got %v with colour %d %d", picture.Bounds(), red>>8, green>>8)
	}
	if _, err := camera.SaveLive(project, "en_directo"); !errors.Is(err, domain.ErrAliasTaken) {
		t.Errorf("same name again: got %v, want ErrAliasTaken", err)
	}
}

func TestPhotoWhileLiveNeedsAFrame(t *testing.T) {
	preview := NewPreviewService(t.TempDir(), func() {})
	camera := NewCaptureService(newService(), preview, t.TempDir())
	if _, err := camera.SaveLive(t.TempDir(), "pronto"); !errors.Is(err, ErrCameraStarting) {
		t.Errorf("got %v, want ErrCameraStarting", err)
	}
}

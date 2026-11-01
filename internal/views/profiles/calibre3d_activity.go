package profiles

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
)

// Measures of the window around the activity panel, in pixels, to size the camera view.
const (
	sideColumnWidth = 361 // the cards' column and its separator
	chromeHeight    = 76  // the toolbar, the status bar and their separators
	panelPadding    = 16
	minLogHeight    = 120 // the camera's messages keep at least this much under the picture
)

// drawActivity shows the last thing launched from this profile: a run's output, or the live camera.
func (profile *Calibre3D) drawActivity(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(panelPadding), Gap: 12}
	context.Box(layout, func() {
		if profile.result != nil {
			profile.drawResultStrip(context)
		}
		switch {
		case profile.activity.live:
			context.Text(profile.activity.title, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			profile.drawLive(context)
		case profile.activity.run != nil:
			context.Text(profile.activity.title, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			profile.drawRunActivity(context)
		default:
			context.Text("Actividad", ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			context.Caption("Aquí verás la cámara, la instalación del entorno o las fotos que guardes.")
		}
	})
}

// drawLive shows the camera's picture while it is open, and its messages under it.
func (profile *Calibre3D) drawLive(context *ui.Context) {
	if profile.Preview.Running(profile.project) {
		width, height := cameraSize(context, profile.reservedHeight())
		latest := func(use func(frame domain.Frame, sequence uint64)) bool {
			return profile.Preview.WithLatest(profile.project, use)
		}
		profile.view.Draw(context, latest, ui.Layout{Width: ui.Fixed(width), Height: ui.Fixed(height)})
	}
	components.OutputLines(context, "camera", profile.Preview.Log(profile.project), 0)
}

// cameraSize fits a 16:9 picture in the panel's width, leaving room for the messages under it and
// for reserved pixels above it.
func cameraSize(context *ui.Context, reserved float32) (width, height float32) {
	windowWidth, windowHeight := context.Size()
	width = max(windowWidth-sideColumnWidth-2*panelPadding, 0)
	height = width * 9 / 16
	room := windowHeight - chromeHeight - 2*panelPadding - 2*12 - 30 - minLogHeight - reserved
	if height > room {
		height = max(room, 0)
		width = height * 16 / 9
	}
	return width, height
}

// pictureLogHeight is how tall the output stays when a step's pictures are shown under it.
const pictureLogHeight = 170

// drawRunActivity shows a run's output and, once a step ended well, the pictures it left side by side.
func (profile *Calibre3D) drawRunActivity(context *ui.Context) {
	snapshot := profile.activity.run.Snapshot()
	var pictures []string
	if step := profile.activity.step; step != nil && snapshot.Status == services.Succeeded {
		pictures = stepPictures(profile.project, *step)
	}
	if len(pictures) == 0 {
		components.OutputLines(context, "activity", snapshot.Lines, snapshot.Dropped)
		return
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(pictureLogHeight)}, func() {
		components.OutputLines(context, "activity", snapshot.Lines, snapshot.Dropped)
	})
	windowWidth, windowHeight := context.Size()
	width := (windowWidth - sideColumnWidth - 2*panelPadding - float32(len(pictures)-1)*12) / float32(len(pictures))
	height := windowHeight - chromeHeight - 2*panelPadding - 30 - pictureLogHeight - 2*12 - profile.reservedHeight()
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12}, func() {
		for _, picture := range pictures {
			profile.images.Draw(context, picture, width, height)
		}
	})
}

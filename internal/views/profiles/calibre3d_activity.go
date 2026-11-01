package profiles

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
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
		switch {
		case profile.activity.live:
			context.Text(profile.activity.title, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			profile.drawLive(context)
		case profile.activity.run != nil:
			context.Text(profile.activity.title, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			snapshot := profile.activity.run.Snapshot()
			components.OutputLines(context, "activity", snapshot.Lines, snapshot.Dropped)
		default:
			context.Text("Actividad", ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			context.Caption("Aquí verás la cámara, la instalación del entorno o las fotos que guardes.")
		}
	})
}

// drawLive shows the camera's picture while it is open, and its messages under it.
func (profile *Calibre3D) drawLive(context *ui.Context) {
	if profile.preview.Running(profile.project) {
		width, height := cameraSize(context)
		latest := func(use func(frame domain.Frame, sequence uint64)) bool {
			return profile.preview.WithLatest(profile.project, use)
		}
		profile.view.Draw(context, latest, ui.Layout{Width: ui.Fixed(width), Height: ui.Fixed(height)})
	}
	components.OutputLines(context, "camera", profile.preview.Log(profile.project), 0)
}

// cameraSize fits a 16:9 picture in the panel's width, leaving room for the messages under it.
func cameraSize(context *ui.Context) (width, height float32) {
	windowWidth, windowHeight := context.Size()
	width = max(windowWidth-sideColumnWidth-2*panelPadding, 0)
	height = width * 9 / 16
	room := windowHeight - chromeHeight - 2*panelPadding - 2*12 - 30 - minLogHeight
	if height > room {
		height = max(room, 0)
		width = height * 16 / 9
	}
	return width, height
}

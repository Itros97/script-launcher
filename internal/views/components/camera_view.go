package components

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
)

// FrameSource gives the latest camera frame and its sequence number, which grows with each new frame.
type FrameSource func(use func(frame domain.Frame, sequence uint64)) bool

// CameraView shows a live camera, uploading a frame's pixels only when a new one arrived.
type CameraView struct {
	picture  *draw2d.Image
	sequence uint64
}

// Draw shows the latest frame filling layout, or a note while the first frame is on its way.
func (view *CameraView) Draw(context *ui.Context, latest FrameSource, layout ui.Layout) {
	shown := latest(func(frame domain.Frame, sequence uint64) {
		if view.picture == nil || !sameSize(view.picture, frame) {
			view.picture, view.sequence = draw2d.NewBlankImage(frame.Width, frame.Height), 0
		}
		if sequence == view.sequence {
			return
		}
		copyRGB(view.picture.Pixels.Pix, frame.RGB)
		view.picture.Changed()
		view.sequence = sequence
	})
	if !shown {
		look := ui.Look{Fill: context.Theme.Sunken, Radius: 8}
		layout.AlignX, layout.AlignY = ui.Center, ui.Center
		context.Frame(layout, look, func() { context.Caption("Arrancando la cámara…") })
		return
	}
	context.Image(view.picture, layout, 8)
}

func sameSize(picture *draw2d.Image, frame domain.Frame) bool {
	width, height := picture.Size()
	return width == frame.Width && height == frame.Height
}

// copyRGB turns packed RGB pixels into opaque RGBA ones.
func copyRGB(rgba, rgb []byte) {
	for source, target := 0, 0; source+2 < len(rgb) && target+3 < len(rgba); source, target = source+3, target+4 {
		rgba[target], rgba[target+1], rgba[target+2], rgba[target+3] = rgb[source], rgb[source+1], rgb[source+2], 255
	}
}

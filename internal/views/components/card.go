package components

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Card is a bordered surface with a small title above its content.
func Card(context *ui.Context, title string, content func()) {
	theme := context.Theme
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: 10}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(14), Gap: 8}, look, func() {
		context.Text(title, ui.TextStyle{Weight: draw2d.Bold})
		content()
	})
}

// Badge is an icon and a word in one colour, for a status.
func Badge(context *ui.Context, icon, text string, colour draw2d.Color) {
	context.Box(ui.Layout{Gap: 6, AlignY: ui.Center}, func() {
		context.IconSized(icon, 16, colour)
		context.Text(text, ui.TextStyle{Weight: draw2d.Medium, Color: colour})
	})
}

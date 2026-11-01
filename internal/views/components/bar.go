// Package components holds the pieces profiles build their layouts from: the script list, the
// output panel, bars. Each one draws from a session and calls its actions.
package components

import ui "github.com/akrck02/go-ui"

// Bar is a full-width strip painted with the surface colour, its content lined up in the middle.
func Bar(context *ui.Context, height float32, content func()) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.PadXY(12, 0), Gap: 8, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

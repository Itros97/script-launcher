package components

import (
	"strings"

	ui "github.com/akrck02/go-ui"
)

// Paragraph shows text that breaks into lines at the width it is given, word by word.
func Paragraph(context *ui.Context, text string, style ui.TextStyle) {
	space := style.Size
	if space == 0 {
		space = context.Theme.TextSize
	}
	layout := ui.Layout{Width: ui.Grow(), Wrap: true, Gap: space * 0.3, LineGap: ui.NoLineGap}
	context.Box(layout, func() {
		for _, word := range strings.Fields(text) {
			context.Text(word, style)
		}
	})
}

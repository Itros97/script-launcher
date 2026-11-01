package components

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// OutputPanel shows the selected script's name, a summary of its last run, and that run's output.
func OutputPanel(context *ui.Context, current *session.Session) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
	context.Box(layout, func() {
		script, ok := current.Selected()
		if !ok {
			context.Caption("Selecciona un script para ver su salida.")
			return
		}
		run, started := current.Latest(script.ID)
		context.Row(func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
				context.Text(script.Name, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
				context.Caption(ShortPath(script.Path, current.Home()))
			})
			if started {
				context.Caption(runSummary(run.Snapshot()))
			}
		})
		if !started {
			context.Caption("Aún no se ha ejecutado en esta sesión. Pulsa ▶ para ejecutarlo.")
			return
		}
		snapshot := run.Snapshot()
		OutputLines(context, script.ID, snapshot.Lines, snapshot.Dropped)
	})
}

// OutputLines shows output lines in a sunken well that follows new lines until the user scrolls away;
// dropped says how many earlier lines were discarded.
func OutputLines(context *ui.Context, name string, lines []domain.Line, dropped int) {
	theme := context.Theme
	look := ui.Look{Fill: theme.Sunken, Border: theme.Line, BorderWidth: 1, Radius: 8, Clip: true}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, look, func() {
		scroll := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(10), Gap: 1}
		context.ScrollAreaFromEnd("output-"+name, scroll, func() {
			if dropped > 0 {
				context.Caption(fmt.Sprintf("… %d líneas anteriores descartadas", dropped))
			}
			for _, line := range lines {
				context.Text(line.Text, ui.TextStyle{Color: lineColour(context, line.Stream)})
			}
		})
	})
}

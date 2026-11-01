package views

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// drawOutput is the right side: the selected script's header and the output of its last run.
func (app *AppView) drawOutput(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
	context.Box(layout, func() {
		script, ok := app.selected()
		if !ok {
			context.Caption("Select a script to see its output.")
			return
		}
		run, started := app.service.Latest(script.ID)
		app.drawOutputHeader(context, script, run, started)
		if !started {
			context.Caption("Not run yet in this session. Press ▶ to run it.")
			return
		}
		drawLines(context, script.ID, run.Snapshot())
	})
}

// drawOutputHeader names the script and sums up its last run.
func (app *AppView) drawOutputHeader(context *ui.Context, script domain.Script, run *services.Run, started bool) {
	context.Row(func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			context.Text(script.Name, ui.TextStyle{Size: 18, Weight: draw2d.Bold})
			context.Caption(shortPath(script.Path, app.home))
		})
		if started {
			context.Caption(runSummary(run.Snapshot()))
		}
	})
}

// drawLines shows the output in a sunken well that follows new lines until the user scrolls away.
func drawLines(context *ui.Context, scriptID string, snapshot services.RunSnapshot) {
	theme := context.Theme
	look := ui.Look{Fill: theme.Sunken, Border: theme.Line, BorderWidth: 1, Radius: 8, Clip: true}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, look, func() {
		scroll := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(10), Gap: 1}
		context.ScrollAreaFromEnd("output-"+scriptID, scroll, func() {
			if snapshot.Dropped > 0 {
				context.Caption(fmt.Sprintf("… %d earlier lines dropped", snapshot.Dropped))
			}
			for _, line := range snapshot.Lines {
				context.Text(line.Text, ui.TextStyle{Color: lineColour(context, line.Stream)})
			}
		})
	})
}

// lineColour tells streams apart: errors in red, the launcher's own notes in the primary colour.
func lineColour(context *ui.Context, stream domain.Stream) draw2d.Color {
	switch stream {
	case domain.Stderr:
		return context.Theme.Danger
	case domain.System:
		return context.Theme.Primary
	default:
		return context.Theme.Ink
	}
}

// runSummary is the one-line state of a run for the header.
func runSummary(snapshot services.RunSnapshot) string {
	duration := formatDuration(snapshot.Duration)
	switch snapshot.Status {
	case services.Running:
		return fmt.Sprintf("running with %s · %s", snapshot.Runner, duration)
	case services.Succeeded:
		return fmt.Sprintf("exit 0 · %s", duration)
	case services.Stopped:
		return fmt.Sprintf("stopped · %s", duration)
	default:
		return fmt.Sprintf("exit %d · %s", snapshot.ExitCode, duration)
	}
}

package views

import (
	"path/filepath"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// drawScriptList is the left column: one row per saved script, or a hint when there are none.
func (app *AppView) drawScriptList(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(340), Height: ui.Grow()}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Box(ui.Layout{Padding: ui.PadXY(14, 10)}, func() { context.Caption("SAVED SCRIPTS") })
		if len(app.scripts) == 0 {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(14, 4), Gap: 4}, func() {
				context.Label("No scripts yet.")
				context.Caption("Use “Add script” to pick one.")
			})
			return
		}
		context.ScrollArea("scripts", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.PadXY(6, 0), Gap: 2}, func() {
			for _, script := range app.scripts {
				app.drawScriptRow(context, script)
			}
		})
	})
}

// drawScriptRow is one script: its name and path, a status icon, and run/stop and remove buttons.
func (app *AppView) drawScriptRow(context *ui.Context, script domain.Script) {
	theme := context.Theme
	context.Push(script.ID)
	defer context.Pop()
	normal := ui.Look{Radius: 8}
	if script.ID == app.selectedID {
		normal.Fill = theme.Primary.WithAlpha(0.12)
	}
	looks := ui.Looks{Normal: normal, Hovered: ui.Look{Fill: theme.Sunken, Radius: 8}, Pressed: ui.Look{Fill: theme.Sunken, Radius: 8}}
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(8, 6), Gap: 6, AlignY: ui.Center}
	run, started := app.service.Latest(script.ID)
	running := started && run.Running()
	available := fileExists(script.Path)
	clicked := context.Clickable("row", layout, looks, func() {
		// Share starts from nothing, so a long path is clipped instead of pushing the buttons out.
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(script.Name, ui.TextStyle{Weight: draw2d.Medium})
			context.Caption(shortPath(filepath.Dir(script.Path), app.home))
		})
		app.drawStatusIcon(context, run, started, available)
		switch {
		case running:
			if context.IconButton("square", "##Stop") {
				app.stopScript(script)
			}
			context.Tooltip("Stop")
		case available:
			if context.IconButton("play", "##Run") {
				app.runScript(script)
			}
			context.Tooltip("Run")
		}
		if context.DangerFlatIconButton("trash-2", "##Remove") {
			app.removeScript(script)
		}
		context.Tooltip("Remove from the list")
	})
	if clicked {
		app.selectedID = script.ID
	}
}

// drawStatusIcon shows a missing file, or how the last run went.
func (app *AppView) drawStatusIcon(context *ui.Context, run *services.Run, started, available bool) {
	theme := context.Theme
	switch {
	case !available:
		context.IconSized("triangle-alert", 16, theme.Warning)
		context.Tooltip("The file is no longer there")
	case !started:
	default:
		icon, colour := statusLook(context, run.Snapshot().Status)
		context.IconSized(icon, 16, colour)
	}
}

// statusLook is the icon and colour that stand for a run status.
func statusLook(context *ui.Context, status services.RunStatus) (string, draw2d.Color) {
	theme := context.Theme
	switch status {
	case services.Running:
		return "loader-circle", theme.Primary
	case services.Succeeded:
		return "circle-check", theme.Positive
	case services.Stopped:
		return "square", theme.InkSoft
	default:
		return "circle-x", theme.Danger
	}
}

package components

import (
	"path/filepath"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// ScriptList is a column with one row per saved script, or a hint when there are none.
func ScriptList(context *ui.Context, current *session.Session, width ui.Sizing) {
	layout := ui.Layout{Direction: ui.Vertical, Width: width, Height: ui.Grow()}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Box(ui.Layout{Padding: ui.PadXY(14, 10)}, func() { context.Caption("SCRIPTS GUARDADOS") })
		if len(current.Scripts()) == 0 {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(14, 4), Gap: 4}, func() {
				context.Label("Aún no hay scripts.")
				context.Caption("Usa «Añadir script» para elegir uno.")
			})
			return
		}
		context.ScrollArea("scripts", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.PadXY(6, 0), Gap: 2}, func() {
			for _, script := range current.Scripts() {
				ScriptRow(context, current, script)
			}
		})
	})
}

// ScriptRow is one script: its name and folder, a status icon, and run/stop and remove buttons.
func ScriptRow(context *ui.Context, current *session.Session, script domain.Script) {
	theme := context.Theme
	context.Push(script.ID)
	defer context.Pop()
	normal := ui.Look{Radius: 8}
	if current.IsSelected(script.ID) {
		normal.Fill = theme.Primary.WithAlpha(0.12)
	}
	looks := ui.Looks{Normal: normal, Hovered: ui.Look{Fill: theme.Sunken, Radius: 8}, Pressed: ui.Look{Fill: theme.Sunken, Radius: 8}}
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(8, 6), Gap: 6, AlignY: ui.Center}
	clicked := context.Clickable("row", layout, looks, func() {
		// Share starts from nothing, so a long path is clipped instead of pushing the buttons out.
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(script.Name, ui.TextStyle{Weight: draw2d.Medium})
			context.Caption(ShortPath(filepath.Dir(script.Path), current.Home()))
		})
		RunButtons(context, current, script)
		if context.DangerFlatIconButton("trash-2", "##Quitar") {
			current.Remove(script)
		}
		context.Tooltip("Quitar de la lista")
	})
	if clicked {
		current.Select(script.ID)
	}
}

// RunButtons shows how the last run went and a run or stop button; a missing file gets a warning instead.
func RunButtons(context *ui.Context, current *session.Session, script domain.Script) {
	run, started := current.Latest(script.ID)
	switch {
	case !fileExists(script.Path):
		context.IconSized("triangle-alert", 16, context.Theme.Warning)
		context.Tooltip("El archivo ya no está")
		return
	case started:
		icon, colour := statusLook(context, run.Snapshot().Status)
		context.IconSized(icon, 16, colour)
	}
	if started && run.Running() {
		if context.IconButton("square", "##Parar") {
			current.Stop(script)
		}
		context.Tooltip("Parar")
		return
	}
	if context.IconButton("play", "##Ejecutar") {
		current.Run(script)
	}
	context.Tooltip("Ejecutar")
}

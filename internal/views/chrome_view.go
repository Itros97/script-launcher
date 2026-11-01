package views

import (
	"fmt"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// bar is a full-width strip painted with the surface colour.
func bar(context *ui.Context, height float32, content func()) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.PadXY(12, 0), Gap: 8, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

// drawToolbar is the top strip: the app's name and the button that adds a script.
func (app *AppView) drawToolbar(context *ui.Context) {
	bar(context, 48, func() {
		context.IconSized("terminal", 18, context.Theme.Primary)
		context.Text("Script Launcher", ui.TextStyle{Size: 15, Weight: draw2d.Bold})
		context.Spacer()
		if context.PrimaryIconButton("plus", "Add script") {
			app.picking = true
		}
		context.Tooltip("Choose a script on disk and save it in the list")
	})
}

// drawStatusBar is the bottom strip: how many scripts are saved and how many are running.
func (app *AppView) drawStatusBar(context *ui.Context) {
	running := 0
	for _, script := range app.scripts {
		if run, ok := app.service.Latest(script.ID); ok && run.Running() {
			running++
		}
	}
	bar(context, 26, func() {
		context.Caption(fmt.Sprintf("%d saved", len(app.scripts)))
		if running > 0 {
			context.Caption(fmt.Sprintf("· %d running", running))
			context.RequestFrame(time.Second) // keeps the elapsed time moving
		}
	})
}

// drawPicker shows go-ui's file picker while the user is adding a script.
func (app *AppView) drawPicker(context *ui.Context) {
	if !app.picking {
		return
	}
	path, closed := context.FilePicker(ui.FilePickerOptions{Title: "Add a script", Start: app.pickerStart()})
	if !closed {
		return
	}
	app.picking = false
	if path != "" {
		app.addScript(path)
	}
}

// drawNotice shows the last notice as a toast until it expires.
func (app *AppView) drawNotice(context *ui.Context) {
	remaining := app.noticeEnd.Sub(app.now())
	if app.notice == "" || remaining <= 0 {
		app.notice = ""
		return
	}
	context.Toast(app.notice, app.warning)
	context.RequestFrame(remaining)
}

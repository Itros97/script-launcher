package views

import (
	"fmt"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/profiles"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// AppView is the whole window: the toolbar with the profile selector, the active profile's layout,
// the status bar, the file picker and the notices.
type AppView struct {
	session  *session.Session
	profiles profiles.Registry
	settings data.SettingsRepository
	active   int
}

// NewAppView opens on the profile saved last time; a corrupt settings file is an error so it is never overwritten.
func NewAppView(current *session.Session, registry profiles.Registry, settings data.SettingsRepository) (*AppView, error) {
	saved, err := settings.Load()
	if err != nil {
		return nil, err
	}
	return &AppView{session: current, profiles: registry, settings: settings, active: registry.IndexOf(saved.Profile)}, nil
}

// Draw describes the window for this frame, top to bottom.
func (app *AppView) Draw(context *ui.Context) {
	context.KeyboardNavigation = ui.Always
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		app.drawToolbar(context)
		context.Separator()
		app.profiles.Profiles()[app.active].Draw(context, app.session)
		context.Separator()
		app.drawStatusBar(context)
	})
	app.drawPicker(context)
	app.drawNotice(context)
}

// drawToolbar is the top strip: the app's name, the profile selector and the button that adds a script.
func (app *AppView) drawToolbar(context *ui.Context) {
	components.Bar(context, 48, func() {
		context.IconSized("terminal", 18, context.Theme.Primary)
		context.Text("Script Launcher", ui.TextStyle{Size: 15, Weight: draw2d.Bold})
		context.Box(ui.Layout{Width: ui.Fixed(12)}, nil)
		context.Box(ui.Layout{Width: ui.Fixed(160)}, func() {
			chosen := app.active
			if context.Select("Perfil", &chosen, app.profiles.Names()) {
				app.switchProfile(chosen)
			}
		})
		context.Tooltip("Perfil: la distribución de la ventana")
		context.Spacer()
		if context.PrimaryIconButton("plus", "Añadir script") {
			app.session.OpenPicker()
		}
		context.Tooltip("Elige un script del disco y guárdalo en la lista")
	})
}

// switchProfile shows another layout and remembers it for the next time the app opens.
func (app *AppView) switchProfile(index int) {
	if leaving, ok := app.profiles.Profiles()[app.active].(profiles.Leaver); ok && index != app.active {
		leaving.Leave(app.session)
	}
	app.active = index
	app.session.ClosePicker()
	profile := app.profiles.Profiles()[index].ID()
	if err := app.settings.Update(func(settings *data.Settings) { settings.Profile = profile }); err != nil {
		app.session.Warn(err)
	}
}

// drawStatusBar is the bottom strip: how many scripts are saved and how many are running.
func (app *AppView) drawStatusBar(context *ui.Context) {
	running := app.session.RunningCount()
	components.Bar(context, 26, func() {
		context.Caption(fmt.Sprintf("%d guardados", len(app.session.Scripts())))
		if running > 0 {
			context.Caption(fmt.Sprintf("· %d en marcha", running))
			context.RequestFrame(time.Second) // keeps the elapsed time moving
		}
	})
}

// drawPicker shows go-ui's file picker while the user is adding a script.
func (app *AppView) drawPicker(context *ui.Context) {
	if !app.session.Picking() {
		return
	}
	path, closed := context.FilePicker(ui.FilePickerOptions{Title: "Añadir un script", Start: app.session.PickerStart()})
	if !closed {
		return
	}
	app.session.ClosePicker()
	if path != "" {
		app.session.Add(path)
	}
}

// drawNotice shows the current notice as a toast until it expires.
func (app *AppView) drawNotice(context *ui.Context) {
	notice, shown := app.session.CurrentNotice()
	if !shown {
		return
	}
	context.Toast(notice.Text, notice.Warning)
	context.RequestFrame(notice.Ends.Sub(app.session.Now()))
}

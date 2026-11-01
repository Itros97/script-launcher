package profiles

import (
	"path/filepath"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// drawProjectCard shows the chosen folder and the button that changes it.
func (profile *Calibre3D) drawProjectCard(context *ui.Context, current *session.Session) {
	components.Card(context, "Proyecto", func() {
		if profile.project == "" {
			context.Caption("Aún no has elegido la carpeta de calibre3d.")
		} else {
			context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
				context.Text(filepath.Base(profile.project), ui.TextStyle{Weight: draw2d.Medium})
				context.Caption(components.ShortPath(filepath.Dir(profile.project), current.Home()))
			})
		}
		if context.IconButton("folder-open", "Elegir carpeta") {
			profile.choosing = true
		}
	})
}

// drawFolderPicker shows the folder picker while the user is choosing the project.
func (profile *Calibre3D) drawFolderPicker(context *ui.Context, current *session.Session) {
	if !profile.choosing {
		return
	}
	folder, closed := context.FilePicker(ui.FilePickerOptions{Title: "Carpeta de calibre3d", Folder: true, Start: profile.project})
	if !closed {
		return
	}
	profile.choosing = false
	if folder != "" {
		profile.choose(current, folder)
	}
}

// choose saves a new project folder and checks it.
func (profile *Calibre3D) choose(current *session.Session, folder string) {
	if err := profile.settings.Update(func(settings *data.Settings) { settings.Calibre3DProject = folder }); err != nil {
		current.Warn(err)
		return
	}
	profile.project = folder
	profile.check(current)
}

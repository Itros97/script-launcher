package profiles

import (
	"strconv"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// drawPythonCard shows whether the Python the environment is made with is installed, and the button
// that installs it; on a computer with no Python to download it shows nothing.
func (profile *Calibre3D) drawPythonCard(context *ui.Context, current *session.Session) {
	download, ok := profile.Env.Download()
	if !ok {
		return
	}
	size := "unos " + strconv.Itoa(download.SizeMB) + " MB"
	hint := ui.TextStyle{Size: context.Theme.SmallTextSize, Color: context.Theme.InkSoft}
	components.Card(context, "Python", func() {
		switch {
		case profile.pythonInstalling() || (profile.setupRunning() && !profile.pythonReady):
			components.Badge(context, "loader-circle", "Instalando…", context.Theme.Primary)
			components.Paragraph(context, "Descargando Python "+download.Version+" ("+size+"). Solo pasa una vez.", hint)
		case profile.pythonReady:
			components.Badge(context, "circle-check", "Python "+download.Version+" listo", context.Theme.Positive)
			components.Paragraph(context, "Script Launcher lo usa para preparar el entorno.", hint)
		default:
			components.Badge(context, "circle-x", "Sin Python", context.Theme.Danger)
			components.Paragraph(context, "Pulsa el botón: se descarga e instala solo, sin contraseña ("+size+").", hint)
			if context.PrimaryIconButton("download", "Instalar Python "+download.Version) {
				profile.installPython(current)
			}
		}
	})
}

// ---- actions ----

// installPython starts installing Python and follows it on the right.
func (profile *Calibre3D) installPython(current *session.Session) {
	run, err := profile.Env.InstallPython()
	if err != nil {
		current.Warn(err)
		return
	}
	profile.show("Instalación de Python", run)
}

// checkAfterPythonInstall looks at the disk again once, when a Python installation ends.
func (profile *Calibre3D) checkAfterPythonInstall(current *session.Session) {
	run, ok := profile.Env.LatestPythonInstall()
	if !ok || run.Running() || run == profile.pythonChecked {
		return
	}
	profile.pythonChecked = run
	profile.check(current)
}

// pythonInstalling tells whether Python is being installed now.
func (profile *Calibre3D) pythonInstalling() bool {
	run, ok := profile.Env.LatestPythonInstall()
	return ok && run.Running()
}

package profiles

import (
	"os"
	"path/filepath"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// drawEnvCard shows the environment's status and the buttons that prepare and check it.
func (profile *Calibre3D) drawEnvCard(context *ui.Context, current *session.Session) {
	components.Card(context, "Entorno", func() {
		icon, word, hint, colour := envLook(context, profile.status, profile.Env.Module())
		running := profile.setupRunning()
		if running {
			icon, word, hint, colour = "loader-circle", "Preparando…", "Instalando lo necesario. Puede tardar unos minutos.", context.Theme.Primary
		}
		if profile.status == domain.Incomplete && profile.reason != "" && !running {
			hint = profile.reason
		}
		components.Badge(context, icon, word, colour)
		components.Paragraph(context, hint, ui.TextStyle{Size: context.Theme.SmallTextSize, Color: context.Theme.InkSoft})
		if profile.status == domain.NoProject || running {
			return
		}
		if !profile.hasReqs {
			missing := "Falta " + domain.RequirementsFile + ": descarga la plantilla."
			context.Text(missing, ui.TextStyle{Size: context.Theme.SmallTextSize, Color: context.Theme.Warning})
		}
		context.Row(func() {
			if profile.setupButton(context) {
				profile.setup(current)
			}
			context.Tooltip("Crea .venv e instala requirements.txt")
			if context.IconButton("refresh-cw", "Comprobar") {
				profile.check(current)
			}
		})
		if context.FlatIconButton("file-down", "Descargar plantilla de requirements.txt") {
			profile.saveTemplate(current)
		}
		context.Tooltip("Guarda en la carpeta del proyecto un requirements.txt con lo que necesita calibre3d")
	})
}

// setupButton is the main call to action until the environment is ready, and a quieter one after.
func (profile *Calibre3D) setupButton(context *ui.Context) bool {
	if profile.status == domain.Ready {
		return context.IconButton("download", "Volver a preparar")
	}
	return context.PrimaryIconButton("download", "Preparar entorno")
}

// envLook is the icon, word, hint and colour that stand for an environment status.
func envLook(context *ui.Context, status domain.EnvStatus, module string) (icon, word, hint string, colour draw2d.Color) {
	theme := context.Theme
	switch status {
	case domain.NoProject:
		return "circle-x", "Sin proyecto", "Elige la carpeta del proyecto " + module + ".", theme.Danger
	case domain.NoVenv:
		return "circle-x", "Sin entorno", "Pulsa «Preparar entorno» para instalar todo lo necesario.", theme.Danger
	case domain.Incomplete:
		return "triangle-alert", "Incompleto", "El entorno existe pero " + module + " no se puede cargar. Vuelve a prepararlo.", theme.Warning
	default:
		return "circle-check", "Listo", "Todo preparado para ejecutar " + module + ".", theme.Positive
	}
}

// ---- actions ----

// check looks at the environment again and shows a problem if the check itself failed.
func (profile *Calibre3D) check(current *session.Session) {
	result, err := profile.Env.Check(profile.project)
	profile.status, profile.reason = result.Status, result.Reason
	_, missing := os.Stat(filepath.Join(profile.project, domain.RequirementsFile))
	profile.hasReqs = missing == nil
	if err != nil {
		current.Warn(err)
	}
}

// saveTemplate writes the requirements template into the project, never over an existing file.
func (profile *Calibre3D) saveTemplate(current *session.Session) {
	path, err := services.SaveRequirementsTemplate(profile.project)
	if err != nil {
		current.Warn(err)
		return
	}
	current.Notify("Plantilla guardada: "+filepath.Base(path), false)
	profile.check(current)
}

// setup starts preparing the environment and follows it on the right.
func (profile *Calibre3D) setup(current *session.Session) {
	run, err := profile.Env.Setup(profile.project)
	if err != nil {
		current.Warn(err)
		return
	}
	profile.show("Instalación", run)
}

// checkAfterSetup checks the environment once, when a setup run ends.
func (profile *Calibre3D) checkAfterSetup(current *session.Session) {
	run, ok := profile.Env.LatestSetup(profile.project)
	if !ok || run.Running() || run == profile.checked {
		return
	}
	profile.checked = run
	profile.check(current)
}

// setupRunning tells whether the project's environment is being prepared now.
func (profile *Calibre3D) setupRunning() bool {
	run, ok := profile.Env.LatestSetup(profile.project)
	return ok && run.Running()
}

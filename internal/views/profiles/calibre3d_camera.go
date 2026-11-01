package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// photoCapture is a photo being saved, kept to announce how it went once.
type photoCapture struct {
	alias     string
	run       *services.Run
	announced bool
}

// drawCameraCard offers to save a photo once the environment is ready.
func (profile *Calibre3D) drawCameraCard(context *ui.Context, current *session.Session) {
	components.Card(context, "Cámara", func() {
		switch {
		case profile.status != domain.Ready:
			context.Caption("Prepara el entorno para usar la cámara.")
		case profile.stepRunning():
			context.Caption("Espera a que termine la medida.")
		case profile.capturing():
			components.Badge(context, "loader-circle", "Guardando «"+profile.capture.alias+"»…", context.Theme.Primary)
		default:
			context.Caption("Guarda lo que ve la cámara ahora, en " + domain.PhotosDir + ".")
			context.Row(func() {
				profile.drawLiveToggle(context, current)
				if context.PrimaryIconButton("camera", "Guardar imagen") {
					profile.askName(photoName)
				}
			})
		}
	})
}

// drawLiveToggle opens or closes the live camera in the activity panel.
func (profile *Calibre3D) drawLiveToggle(context *ui.Context, current *session.Session) {
	if profile.Preview.Running(profile.project) {
		if context.IconButton("video-off", "Parar cámara") {
			profile.Leave(current)
		}
		return
	}
	if context.IconButton("video", "Ver cámara") {
		profile.startPreview(current)
	}
}

// ---- actions ----

// startPreview opens the live camera and shows it on the right.
func (profile *Calibre3D) startPreview(current *session.Session) {
	if err := profile.Preview.Start(profile.project); err != nil {
		current.Warn(err)
		return
	}
	profile.activity = activity{title: "Cámara", live: true}
}

// save keeps the frame on screen while the camera is live, and otherwise captures one and follows it on the right.
func (profile *Calibre3D) save(current *session.Session, alias string) {
	if profile.Camera.Live(profile.project) {
		if _, err := profile.Camera.SaveLive(profile.project, alias); err != nil {
			current.Warn(err)
			return
		}
		current.Notify("Imagen guardada: "+alias+".png", false)
		return
	}
	run, err := profile.Camera.Capture(profile.project, alias)
	if err != nil {
		current.Warn(err)
		return
	}
	profile.capture = photoCapture{alias: alias, run: run}
	profile.show("Foto «"+alias+"»", run)
}

// capturing tells whether a photo is being saved now.
func (profile *Calibre3D) capturing() bool {
	return profile.capture.run != nil && profile.capture.run.Running()
}

// announceCapture says once how the last photo went, when it ends.
func (profile *Calibre3D) announceCapture(current *session.Session) {
	capture := &profile.capture
	if capture.run == nil || capture.announced || capture.run.Running() {
		return
	}
	capture.announced = true
	if capture.run.Snapshot().Status == services.Succeeded {
		current.Notify("Imagen guardada: "+capture.alias+".png", false)
		return
	}
	current.Notify("No se pudo guardar la imagen: mira el detalle a la derecha.", true)
}

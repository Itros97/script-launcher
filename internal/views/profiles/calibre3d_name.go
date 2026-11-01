package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// nameKind is what the name dialog is asking a name for.
type nameKind int

const (
	noName      nameKind = iota // the dialog is closed
	photoName                   // a photo to save
	measureName                 // a piece to measure
)

// askName opens the name dialog, empty, for a photo or a piece.
func (profile *Calibre3D) askName(kind nameKind) {
	profile.naming, profile.opening, profile.name = kind, true, ""
}

// drawNameDialog asks for the name and, once it is accepted, saves the photo or measures the piece.
func (profile *Calibre3D) drawNameDialog(context *ui.Context, current *session.Session) {
	if profile.naming == noName {
		return
	}
	title, placeholder, validate := "Guardar imagen", "por ejemplo: pieza1", func(name string) error { return profile.Camera.Validate(profile.project, name) }
	if profile.naming == measureName {
		title, placeholder, validate = "Medir una pieza", "por ejemplo: muestra_001", func(name string) error { return profile.Calibre.Validate(profile.project, name) }
	}
	confirmed, closed := components.NameDialog(context, title, placeholder, &profile.name, profile.opening, validate)
	profile.opening = false
	kind := profile.naming
	if closed {
		profile.naming = noName
	}
	switch {
	case confirmed && kind == measureName:
		profile.measure(current, profile.name)
	case confirmed:
		profile.save(current, profile.name)
	}
}

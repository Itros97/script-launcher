package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// stepRun is a calibre3d step started from the profile, kept to handle its end once.
type stepRun struct {
	step    domain.CalibreStep
	name    string
	run     *services.Run
	handled bool
}

// planeState is the project's reference plane as last read; loaded says whether it was read at all.
type planeState struct {
	info   domain.PlaneInfo
	saved  bool
	loaded bool
}

// drawMeasureCard lists the steps in order, each with what to do before pressing it.
func (profile *Calibre3D) drawMeasureCard(context *ui.Context, current *session.Session) {
	components.Card(context, "Medir", func() {
		if profile.status != domain.Ready {
			context.Caption("Prepara el entorno para medir.")
			return
		}
		if profile.busy() {
			components.Badge(context, "loader-circle", "Espera a que termine…", context.Theme.Primary)
			return
		}
		profile.loadPlane(current)
		profile.loadCalibration(current)
		hint := context.Theme.InkSoft
		if profile.drawStep(context, "scan", "1. Plano", planeHint(profile.plane), hint) {
			profile.runStep(current, domain.StepPlane, "")
		}
		if profile.drawStep(context, "crosshair", "2. Centrar", "Pon una moneda en la cruz.", hint) {
			profile.runStep(current, domain.StepCenter, "")
		}
		patternHint, measureHint, measureColour := calibrationHint(profile.calibration), "Pon la pieza dentro del círculo.", hint
		if !profile.plane.saved {
			patternHint, measureHint, measureColour = "Haz primero el plano.", "Haz primero el plano.", context.Theme.Warning
		}
		if profile.drawStep(context, "scale", "3. Patrón", patternHint, measureColour) {
			profile.askPattern()
		}
		if profile.drawStep(context, "ruler", "4. Medir", measureHint, measureColour) {
			profile.askName(measureName)
		}
		if context.FlatIconButton("flask-conical", "Demo sin cámara") {
			profile.runStep(current, domain.StepDemo, "")
		}
		context.Tooltip("Mide una pieza simulada: sirve para probar sin la cámara")
	})
}

// drawStep is one step: its button and, beside it, what to do first.
func (profile *Calibre3D) drawStep(context *ui.Context, icon, label, hint string, colour components.Colour) bool {
	pressed := false
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 10, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Fixed(118)}, func() { pressed = context.IconButton(icon, label) })
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			components.Paragraph(context, hint, ui.TextStyle{Size: context.Theme.SmallTextSize, Color: colour})
		})
	})
	return pressed
}

// planeHint says whether the plane is done, when and how noisy, or what to do to make it.
func planeHint(plane planeState) string {
	if !plane.saved {
		return "Deja el tapete vacío."
	}
	return "Hecho " + components.FormatDate(plane.info.Date) + " · ruido " + components.FormatMM(plane.info.NoiseMM) + " mm"
}

// ---- actions ----

// measure starts measuring a named piece.
func (profile *Calibre3D) measure(current *session.Session, name string) {
	profile.runStep(current, domain.StepMeasure, name)
}

// runStep starts a calibre3d step and follows it on the right; the live camera is closed first.
func (profile *Calibre3D) runStep(current *session.Session, step domain.CalibreStep, name string) {
	profile.startStep(current, step, name, func() (*services.Run, bool, error) {
		return profile.Calibre.Run(profile.project, step, name)
	})
}

// startStep starts a step with start, which closes the live camera first, and follows it on the right.
func (profile *Calibre3D) startStep(current *session.Session, step domain.CalibreStep, name string, start func() (*services.Run, bool, error)) {
	run, closed, err := start()
	if closed {
		current.Notify("He parado la vista de la cámara para medir.", false)
	}
	if err != nil {
		current.Warn(err)
		return
	}
	profile.step = stepRun{step: step, name: name, run: run}
	profile.follow(activity{title: step.Title(name), run: run, step: &profile.step})
}

// stepRunning tells whether a calibre3d step is running now.
func (profile *Calibre3D) stepRunning() bool {
	return profile.step.run != nil && profile.step.run.Running()
}

// finishStep handles a step's end once: reloads the plane and the calibration, and reads and shows a measurement.
func (profile *Calibre3D) finishStep(current *session.Session) {
	step := &profile.step
	if step.run == nil || step.handled || step.run.Running() {
		return
	}
	step.handled = true
	if step.run.Snapshot().Status != services.Succeeded {
		current.Notify("No se pudo completar «"+step.step.Title(step.name)+"»: mira el detalle a la derecha.", true)
		return
	}
	profile.plane.loaded, profile.calibration.loaded = false, false
	if step.step == domain.StepPattern {
		current.Notify("Patrón «"+step.name+"» guardado: las próximas medidas usarán la corrección.", false)
		return
	}
	if step.step != domain.StepMeasure && step.step != domain.StepDemo {
		return
	}
	measurement, err := profile.Calibre.LoadMeasurement(profile.project, step.step, step.name)
	if err != nil {
		current.Warn(err)
		return
	}
	profile.result = &measured{measurement: measurement, step: step.step, name: step.name}
	profile.library.loaded = false
	current.Notify("Medida lista: "+step.step.Title(step.name), false)
}

// loadPlane reads the plane once, and again after each step.
func (profile *Calibre3D) loadPlane(current *session.Session) {
	if profile.plane.loaded {
		return
	}
	info, saved, err := profile.Calibre.LoadPlane(profile.project)
	profile.plane = planeState{info: info, saved: saved, loaded: true}
	if err != nil {
		current.Warn(err)
	}
}

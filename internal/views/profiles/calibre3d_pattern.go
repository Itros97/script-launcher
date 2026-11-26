package profiles

import (
	"fmt"
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// The pattern dialog's fields, named so the first can be given the keyboard and each told apart.
const (
	patternNameField      = "Nombre del patrón"
	patternLengthField    = "Largo real (mm)"
	patternWidthField     = "Ancho real (mm)"
	patternThicknessField = "Espesor real (mm)"
)

// calibrationState is the project's calibration as last read; saved says whether any pattern was measured.
type calibrationState struct {
	info   domain.Calibration
	saved  bool
	loaded bool
}

// patternForm is the pattern dialog: whether it is shown and what is typed in it.
type patternForm struct {
	open, opening                  bool
	name, length, width, thickness string
}

// askPattern opens the pattern dialog, empty.
func (profile *Calibre3D) askPattern() {
	profile.pattern = patternForm{open: true, opening: true}
}

// drawPatternDialog asks for the pattern's name and real sizes and, once they are accepted, measures it.
func (profile *Calibre3D) drawPatternDialog(context *ui.Context, current *session.Session) {
	form := &profile.pattern
	if !form.open {
		return
	}
	confirmed := false
	dismissed := context.Sheet("Medir un patrón", ui.SheetOptions{}, func() {
		context.Caption("Una pieza de tamaño conocido: mídela antes con el calibre de mano.")
		if form.opening {
			context.FocusField(patternNameField)
		}
		submitted := context.EntryField(patternNameField, &form.name, ui.TextFieldOptions{Placeholder: "por ejemplo: bloque_25"})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8}, func() {
			for _, size := range []struct {
				field, label string
				text         *string
			}{{patternLengthField, "Largo", &form.length}, {patternWidthField, "Ancho", &form.width}, {patternThicknessField, "Espesor", &form.thickness}} {
				context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
					context.Caption(size.label + " real (mm)")
					submitted = context.EntryField(size.field, size.text, ui.TextFieldOptions{Placeholder: "0,00"}) || submitted
				})
			}
		})
		pattern, problem := domain.ParsePattern(form.name, form.length, form.width, form.thickness)
		drawPatternProblem(context, problem, form.name == "")
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				form.open = false
			}
			if (context.PrimaryButton("Medir patrón") || submitted) && problem == nil {
				confirmed, form.open = true, false
				profile.runPattern(current, pattern)
			}
		})
	})
	form.opening = false
	if dismissed && !confirmed {
		form.open = false
	}
}

// drawPatternProblem says why the pattern cannot be measured yet: quietly while no name is typed, in red after.
func drawPatternProblem(context *ui.Context, problem error, empty bool) {
	if problem == nil {
		context.Caption(" ")
		return
	}
	colour := context.Theme.Danger
	if empty {
		colour = context.Theme.InkQuiet
	}
	context.Text(session.Describe(problem), ui.TextStyle{Size: context.Theme.SmallTextSize, Color: colour})
}

// runPattern measures the pattern and follows it on the right, as any other step.
func (profile *Calibre3D) runPattern(current *session.Session, pattern domain.Pattern) {
	profile.startStep(current, domain.StepPattern, pattern.Name, func() (*services.Run, bool, error) {
		return profile.Calibre.RunPattern(profile.project, pattern)
	})
}

// loadCalibration reads the calibration once, and again after each step.
func (profile *Calibre3D) loadCalibration(current *session.Session) {
	if profile.calibration.loaded {
		return
	}
	info, saved, err := profile.Calibre.LoadCalibration(profile.project)
	profile.calibration = calibrationState{info: info, saved: saved, loaded: true}
	if err != nil {
		current.Warn(err)
	}
}

// calibrationHint says how many patterns the calibration has and the correction they give, or what to do to make one.
func calibrationHint(calibration calibrationState) string {
	if !calibration.saved {
		return "Sin calibrar: mide una pieza de tamaño conocido."
	}
	correction := calibration.info.Correction
	patterns := "1 patrón"
	if calibration.info.Patterns != 1 {
		patterns = fmt.Sprintf("%d patrones", calibration.info.Patterns)
	}
	return patterns + " · L,A " + formatLinear(correction.KXY, correction.OffsetXY) + " · E " + formatLinear(correction.KZ, correction.OffsetZ)
}

// formatLinear writes a linear correction as "×1,0012 +0,98 mm", with a decimal comma.
func formatLinear(k, offset float64) string {
	sign := "+"
	if offset < 0 {
		sign = "−"
	}
	factor := strings.Replace(fmt.Sprintf("%.4f", k), ".", ",", 1)
	return "×" + factor + " " + sign + components.FormatMM(max(offset, -offset)) + " mm"
}

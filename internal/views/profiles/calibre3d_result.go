package profiles

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/components"
)

// resultStripHeight is the room the result strip takes above the activity, gap included.
const resultStripHeight = 92

// measured is the last measurement read, and the step and name that find its pictures.
type measured struct {
	measurement domain.Measurement
	step        domain.CalibreStep
	name        string
}

// drawResultStrip shows the last measurement above the activity, so it stays in sight whatever runs next.
func (profile *Calibre3D) drawResultStrip(context *ui.Context) {
	measurement := profile.result.measurement
	theme := context.Theme
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: 10}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(16, 10), Gap: 4}, look, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 24, AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
				context.Text("Resultado · "+measurement.Name, ui.TextStyle{Weight: draw2d.Bold})
				context.Caption(components.FormatDate(measurement.Date))
			})
			for _, size := range []struct {
				label string
				value float64
			}{{"Largo", measurement.Length}, {"Ancho", measurement.Width}, {"Espesor", measurement.Thickness}, {"Feret máx.", measurement.Feret}} {
				context.Box(ui.Layout{Direction: ui.Vertical, Gap: 2}, func() {
					context.Caption(size.label)
					context.Text(components.FormatMM(size.value)+" mm", ui.TextStyle{Size: 18, Weight: draw2d.Bold})
				})
			}
		})
		for _, warning := range measurement.Warnings {
			context.Text("⚠ "+warning, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Warning})
		}
	})
}

// reservedHeight is the room the result strip takes in the activity panel, if shown.
func (profile *Calibre3D) reservedHeight() float32 {
	if profile.result == nil {
		return 0
	}
	return resultStripHeight + 16*float32(len(profile.result.measurement.Warnings))
}

// stepPictures are the pictures a finished step left, to show under its output.
func stepPictures(project string, step stepRun) []string {
	switch step.step {
	case domain.StepCenter:
		return []string{step.step.ResultFile(project, step.name, "_vista")}
	case domain.StepMeasure, domain.StepDemo:
		return []string{step.step.ResultFile(project, step.name, "_mapa"), step.step.ResultFile(project, step.name, "_vista")}
	default:
		return nil
	}
}

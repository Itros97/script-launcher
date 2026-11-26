package views

import (
	"os"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/Itros97/script-launcher/internal/domain"
)

// typePattern fills the open pattern dialog field by field, going on with Tab.
func typePattern(app testApp, fields ...string) {
	for index, text := range fields {
		if index > 0 {
			app.window.PressKey(input.KeyTab)
		}
		app.window.Type(text)
		app.window.Frame()
	}
}

func TestThePatternStepSaysTheProjectIsNotCalibrated(t *testing.T) {
	app := openMeasuring(t, measuringProject(t))
	click(t, app.window, "1. Plano")
	waitWhile(t, app, "the plane", func() bool { return !shows(app, "Sin calibrar: mide una pieza de tamaño conocido.") })
}

func TestMeasuringAPatternUpdatesTheCalibration(t *testing.T) {
	project := measuringProject(t)
	app := openMeasuring(t, project)
	click(t, app.window, "1. Plano")
	waitWhile(t, app, "the plane", func() bool { return !shows(app, "Sin calibrar") })
	click(t, app.window, "3. Patrón")
	typePattern(app, "bloque_60", "60", "40,05", "20")
	app.window.PressKey(input.KeyEnter)
	waitForText(t, app.window, "Patrón «bloque_60» guardado: las próximas medidas usarán la corrección.")
	if !shows(app, "1 patrón · L,A ×1,0000 +1,00 mm · E ×1,0000 +0,00 mm") {
		t.Errorf("the calibration is not shown; window shows %q", app.window.Texts())
	}
	if _, err := os.Stat(domain.StepPattern.ResultFile(project, "bloque_60", "_vista")); err != nil {
		t.Errorf("no view: %v", err)
	}
}

func TestAPatternWithoutItsSizesIsNotMeasured(t *testing.T) {
	project := measuringProject(t)
	app := openMeasuring(t, project)
	click(t, app.window, "3. Patrón")
	typePattern(app, "bloque_60", "60", "cuarenta")
	app.window.PressKey(input.KeyEnter)
	app.window.Frame()
	if !shows(app, "Escribe cada medida real en mm, mayor que 0 (con coma o punto).") {
		t.Errorf("the problem is not shown; window shows %q", app.window.Texts())
	}
	if _, err := os.Stat(domain.CalibrationFile(project)); err == nil {
		t.Error("the pattern was measured")
	}
}

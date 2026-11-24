package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
)

// boundFieldWidth is the width of each "mín" or "máx" field.
const boundFieldWidth = 64

// axes are the three sizes the library filters by, in the order of the filter's bounds.
var axes = [...]string{"Largo", "Ancho", "Espesor"}

// filterTexts is the library's filter as typed: the name and, per axis, its minimum and maximum.
type filterTexts struct {
	name   string
	bounds [len(axes)][2]string
}

// parse turns the typed filter into a domain filter; valid says, per bound, whether it was a
// number or empty. An invalid bound does not filter.
func (texts filterTexts) parse() (domain.MeasurementFilter, [len(axes)][2]bool) {
	var ranges [len(axes)]domain.Range
	var valid [len(axes)][2]bool
	for axis, pair := range texts.bounds {
		minimum, minErr := domain.ParseBound(pair[0])
		maximum, maxErr := domain.ParseBound(pair[1])
		ranges[axis] = domain.Range{Min: minimum, Max: maximum}
		valid[axis] = [2]bool{minErr == nil, maxErr == nil}
	}
	return domain.MeasurementFilter{Name: texts.name, Length: ranges[0], Width: ranges[1], Thickness: ranges[2]}, valid
}

// drawFilters shows the name search and the range of each axis; it tells whether any of them changed.
func (profile *Calibre3D) drawFilters(context *ui.Context, valid [len(axes)][2]bool) bool {
	texts := &profile.library.filter
	changed := context.TextFieldWith("Buscar por nombre", &texts.name, ui.TextFieldOptions{Placeholder: "Buscar por nombre"})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 16, AlignY: ui.Center}, func() {
		for axis, label := range axes {
			context.Push(label)
			context.Box(ui.Layout{Gap: 4, AlignY: ui.Center}, func() {
				context.Caption(label)
				for end, placeholder := range [2]string{"mín", "máx"} {
					changed = drawBoundField(context, placeholder, &texts.bounds[axis][end], valid[axis][end]) || changed
				}
			})
			context.Pop()
		}
		if context.FlatButton("Limpiar") {
			*texts, changed = filterTexts{}, true
		}
	})
	if !allValid(valid) {
		context.Text("Escribe los límites en milímetros, por ejemplo 12,5.", ui.TextStyle{Size: context.Theme.SmallTextSize, Color: context.Theme.Danger})
	}
	return changed
}

// drawBoundField is one "mín" or "máx" field, marked in red when its text is not a number.
func drawBoundField(context *ui.Context, name string, text *string, valid bool) bool {
	look := ui.Look{Radius: 8}
	if !valid {
		look.Border, look.BorderWidth = context.Theme.Danger, 2
	}
	changed := false
	context.Frame(ui.Layout{Width: ui.Fixed(boundFieldWidth)}, look, func() {
		changed = context.TextFieldWith(name, text, ui.TextFieldOptions{Placeholder: name})
	})
	return changed
}

// allValid tells whether every bound is a number or empty.
func allValid(valid [len(axes)][2]bool) bool {
	for _, pair := range valid {
		if !pair[0] || !pair[1] {
			return false
		}
	}
	return true
}

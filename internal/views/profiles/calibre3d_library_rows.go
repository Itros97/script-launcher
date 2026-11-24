package profiles

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/components"
)

// sizeColumnWidth is the width of each column of millimetres in the library.
const sizeColumnWidth = 84

// rowButtonsWidth is the room on the right of a row for its warning mark and buttons.
const rowButtonsWidth = 96

// drawLibraryRows shows the column titles and the page's rows, or why there are none.
func (profile *Calibre3D) drawLibraryRows(context *ui.Context, page []domain.Measurement) {
	lib := &profile.library
	if len(page) == 0 {
		message := "Ninguna medición cumple el filtro."
		if len(lib.measurements) == 0 {
			message = "Aún no hay mediciones en este proyecto."
		}
		context.Caption(message)
		return
	}
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(8, 0), Gap: 8}, func() {
		context.Box(ui.Layout{Width: ui.Fixed(context.Theme.CheckSize)}, nil)
		context.Box(ui.Layout{Width: ui.Share(1)}, func() { context.Caption("Nombre") })
		for _, title := range []string{"Largo", "Ancho", "Espesor", "Feret máx."} {
			sizeCell(context, func() { context.Caption(title) })
		}
		context.Box(ui.Layout{Width: ui.Fixed(rowButtonsWidth)}, nil)
	})
	context.ScrollArea("library-rows", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 2}, func() {
		for _, measurement := range page {
			profile.drawLibraryRow(context, measurement)
		}
	})
}

// drawLibraryRow is one measurement: its checkbox, name and date, sizes, and its own export and delete buttons.
func (profile *Calibre3D) drawLibraryRow(context *ui.Context, measurement domain.Measurement) {
	lib := &profile.library
	context.Push(measurement.Name)
	defer context.Pop()
	look := ui.Look{Radius: 8}
	if lib.selected[measurement.Name] {
		look.Fill = context.Theme.Primary.WithAlpha(0.12)
	}
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(8, 4), Gap: 8, AlignY: ui.Center}, look, func() {
		checked := lib.selected[measurement.Name]
		if context.Checkbox("##Seleccionar", &checked) {
			lib.selected[measurement.Name] = checked
		}
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(measurement.Name, ui.TextStyle{Weight: draw2d.Medium})
			context.Caption(components.FormatDate(measurement.Date))
		})
		for _, value := range []float64{measurement.Length, measurement.Width, measurement.Thickness, measurement.Feret} {
			sizeCell(context, func() { context.Text(components.FormatMM(value), ui.TextStyle{}) })
		}
		profile.drawRowButtons(context, measurement)
	})
}

// drawRowButtons are a row's warning mark and its export and delete buttons.
func (profile *Calibre3D) drawRowButtons(context *ui.Context, measurement domain.Measurement) {
	context.Box(ui.Layout{Width: ui.Fixed(rowButtonsWidth), Gap: 4, AlignX: ui.End, AlignY: ui.Center}, func() {
		if len(measurement.Warnings) > 0 {
			context.IconSized("triangle-alert", 16, context.Theme.Warning)
			context.Tooltip(measurement.Warnings[0])
		}
		if context.FlatIconButton("download", "##Exportar") {
			profile.library.exporting = []domain.Measurement{measurement}
		}
		context.Tooltip("Exportar a CSV")
		if context.DangerFlatIconButton("trash-2", "##Borrar") {
			profile.library.deleting = []string{measurement.Name}
		}
		context.Tooltip("Borrar")
	})
}

// drawPager moves between pages and says which one is shown.
func (profile *Calibre3D) drawPager(context *ui.Context, total int) {
	lib := &profile.library
	pages := domain.PageCount(total, domain.LibraryPageSize)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignX: ui.Center, AlignY: ui.Center}, func() {
		if context.FlatIconButton("chevron-left", "Anterior") && lib.page > 0 {
			lib.page--
		}
		context.Caption(fmt.Sprintf("Página %d de %d", lib.page+1, pages))
		if context.FlatIconButton("chevron-right", "Siguiente") && lib.page < pages-1 {
			lib.page++
		}
	})
}

// sizeCell is a cell of the millimetre columns, its content on the right.
func sizeCell(context *ui.Context, content func()) {
	context.Box(ui.Layout{Width: ui.Fixed(sizeColumnWidth), AlignX: ui.End}, content)
}

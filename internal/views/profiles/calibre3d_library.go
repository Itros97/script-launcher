package profiles

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// library is the list of the project's measurements on the right, with its filter, page and selection.
type library struct {
	open         bool
	loaded       bool
	measurements []domain.Measurement
	filter       filterTexts
	page         int
	selected     map[string]bool
	deleting     []string             // the names the delete dialog asks about; nil when it is not shown
	exporting    []domain.Measurement // what the folder picker will export; nil when it is not shown
}

// drawLibraryCard opens the library, for any chosen project: reading it needs no environment.
func (profile *Calibre3D) drawLibraryCard(context *ui.Context) {
	if profile.project == "" {
		return
	}
	components.Card(context, "Biblioteca", func() {
		context.Caption("Todas las medidas guardadas: búscalas, bórralas o expórtalas a CSV.")
		if context.IconButton("library", "Abrir biblioteca") {
			profile.library.open, profile.library.loaded = true, false
		}
	})
}

// drawLibrary shows the filters, one page of the matching measurements, and the page controls.
func (profile *Calibre3D) drawLibrary(context *ui.Context, current *session.Session) {
	profile.loadLibrary(current)
	lib := &profile.library
	filter, valid := lib.filter.parse()
	matching := domain.FilterMeasurements(lib.measurements, filter)
	lib.page = min(lib.page, domain.PageCount(len(matching), domain.LibraryPageSize)-1)
	page := domain.Page(matching, lib.page, domain.LibraryPageSize)
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(panelPadding), Gap: 12}
	context.Box(layout, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
				context.Text("Biblioteca", ui.TextStyle{Size: 18, Weight: draw2d.Bold})
				context.Caption(fmt.Sprintf("%d de %d mediciones", len(matching), len(lib.measurements)))
			})
			if context.FlatIconButton("arrow-left", "Volver") {
				lib.open = false
			}
		})
		if profile.drawFilters(context, valid) {
			lib.page = 0
		}
		profile.drawLibraryActions(context, page, matching)
		profile.drawLibraryRows(context, page)
		profile.drawPager(context, len(matching))
	})
}

// loadLibrary reads the project's measurements once per opening, and again after each change.
func (profile *Calibre3D) loadLibrary(current *session.Session) {
	lib := &profile.library
	if lib.loaded {
		return
	}
	lib.loaded = true
	measurements, unreadable, err := profile.Library.List(profile.project)
	if err != nil {
		current.Warn(err)
	}
	lib.measurements = measurements
	lib.selected = keepExisting(lib.selected, measurements)
	if unreadable > 0 {
		current.Notify(fmt.Sprintf("No he podido leer %d mediciones: sus archivos están dañados.", unreadable), true)
	}
}

// keepExisting is the selection without the names that are no longer in the library.
func keepExisting(selected map[string]bool, measurements []domain.Measurement) map[string]bool {
	kept := map[string]bool{}
	for _, measurement := range measurements {
		if selected[measurement.Name] {
			kept[measurement.Name] = true
		}
	}
	return kept
}

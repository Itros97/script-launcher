package profiles

import (
	"fmt"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// drawLibraryActions are the page's checkbox, the export menu and the delete button for the selection.
func (profile *Calibre3D) drawLibraryActions(context *ui.Context, page, matching []domain.Measurement) {
	lib := &profile.library
	selected := lib.selectedIn(lib.measurements)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		whole := len(page) > 0 && len(lib.selectedIn(page)) == len(page)
		if context.Checkbox("Toda la página", &whole) {
			for _, measurement := range page {
				lib.selected[measurement.Name] = whole
			}
		}
		context.Spacer()
		context.MenuButton("file-spreadsheet", "Exportar CSV", func() {
			if len(selected) > 0 && context.MenuItem("check-square", fmt.Sprintf("Seleccionadas (%d)", len(selected)), "") {
				lib.exporting = selected
			}
			if len(page) > 0 && context.MenuItem("file", fmt.Sprintf("Esta página (%d)", len(page)), "") {
				lib.exporting = page
			}
			if len(matching) > 0 && context.MenuItem("files", fmt.Sprintf("Todas (%d)", len(matching)), "") {
				lib.exporting = matching
			}
		})
		if len(selected) > 0 && context.DangerIconButton("trash-2", fmt.Sprintf("Borrar seleccionadas (%d)", len(selected))) {
			lib.deleting = names(selected)
		}
	})
}

// selectedIn returns the measurements of the list that are selected, in its order.
func (lib *library) selectedIn(measurements []domain.Measurement) []domain.Measurement {
	selected := []domain.Measurement{}
	for _, measurement := range measurements {
		if lib.selected[measurement.Name] {
			selected = append(selected, measurement)
		}
	}
	return selected
}

// drawDeleteDialog asks before deleting, since a deleted measurement cannot be brought back.
func (profile *Calibre3D) drawDeleteDialog(context *ui.Context, current *session.Session) {
	lib := &profile.library
	if lib.deleting == nil {
		return
	}
	question := fmt.Sprintf("¿Borrar «%s»?", lib.deleting[0])
	if len(lib.deleting) > 1 {
		question = fmt.Sprintf("¿Borrar %d mediciones?", len(lib.deleting))
	}
	confirmed := false
	dismissed := context.Sheet("Borrar mediciones", ui.SheetOptions{}, func() {
		context.Label(question)
		components.Paragraph(context, "Se borran también sus imágenes. No se puede deshacer.", ui.TextStyle{Size: context.Theme.SmallTextSize, Color: context.Theme.InkSoft})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				lib.deleting = nil
			}
			confirmed = context.DangerButton("Borrar")
		})
	})
	if confirmed {
		profile.deleteMeasurements(current, lib.deleting)
	}
	if confirmed || dismissed {
		lib.deleting = nil
	}
}

// deleteMeasurements deletes the named measurements and reloads the library.
func (profile *Calibre3D) deleteMeasurements(current *session.Session, names []string) {
	deleted, err := profile.Library.Delete(profile.project, names)
	profile.library.loaded = false
	if err != nil {
		current.Warn(err)
		return
	}
	current.Notify(fmt.Sprintf("Borradas %d mediciones.", deleted), false)
}

// drawExportPicker asks for the folder to export to, starting in the last one used.
func (profile *Calibre3D) drawExportPicker(context *ui.Context, current *session.Session) {
	lib := &profile.library
	if lib.exporting == nil {
		return
	}
	folder, closed := context.FilePicker(ui.FilePickerOptions{Title: "Carpeta para el CSV", Folder: true, Start: profile.exportFolder})
	if !closed {
		return
	}
	measurements := lib.exporting
	lib.exporting = nil
	if folder != "" {
		profile.export(current, folder, measurements)
	}
}

// export writes the CSV in folder and remembers the folder for the next time.
func (profile *Calibre3D) export(current *session.Session, folder string, measurements []domain.Measurement) {
	path, err := profile.Library.Export(folder, measurements)
	if err != nil {
		current.Warn(err)
		return
	}
	profile.exportFolder = folder
	if err := profile.Settings.Update(func(settings *data.Settings) { settings.Calibre3DExportFolder = folder }); err != nil {
		current.Warn(err)
	}
	current.Notify("Exportado a "+components.ShortPath(path, current.Home()), false)
}

// names returns the measurements' names.
func names(measurements []domain.Measurement) []string {
	all := make([]string, len(measurements))
	for index, measurement := range measurements {
		all[index] = measurement.Name
	}
	return all
}

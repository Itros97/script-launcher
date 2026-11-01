package components

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/views/session"
)

// nameField is the dialog's text field, named so it can be given the keyboard when the dialog opens.
const nameField = "Nombre"

// NameDialog asks for a name in a sheet, checking it with validate as it is typed. opening gives
// the field the keyboard; confirmed is true on the step a valid name is accepted (Guardar or
// Enter), and closed on every step the dialog should stop being shown.
func NameDialog(context *ui.Context, title, placeholder string, name *string, opening bool, validate func(string) error) (confirmed, closed bool) {
	dismissed := context.Sheet(title, ui.SheetOptions{}, func() {
		if opening {
			context.FocusField(nameField)
		}
		submitted := context.EntryField(nameField, name, ui.TextFieldOptions{Placeholder: placeholder})
		problem := validate(*name)
		drawProblem(context, problem, *name == "")
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				closed = true
			}
			saved := context.PrimaryButton("Guardar")
			if (saved || submitted) && problem == nil {
				confirmed, closed = true, true
			}
		})
	})
	return confirmed, closed || dismissed
}

// drawProblem says why a name cannot be used: quietly while nothing is typed yet, in red after.
func drawProblem(context *ui.Context, problem error, empty bool) {
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

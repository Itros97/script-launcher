package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// Default is the general layout: the saved scripts on the left, the selected one's output on the right.
type Default struct{}

// ID returns the default profile's ID.
func (Default) ID() domain.ProfileID { return domain.DefaultProfileID }

// Name returns "General".
func (Default) Name() string { return "General" }

// Draw lays out the script list beside the output panel.
func (Default) Draw(context *ui.Context, current *session.Session) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		components.ScriptList(context, current, ui.Fixed(340))
		context.Separator()
		components.OutputPanel(context, current)
	})
}

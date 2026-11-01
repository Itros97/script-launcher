package components

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// statusLook is the icon and colour that stand for a run status.
func statusLook(context *ui.Context, status services.RunStatus) (string, draw2d.Color) {
	theme := context.Theme
	switch status {
	case services.Running:
		return "loader-circle", theme.Primary
	case services.Succeeded:
		return "circle-check", theme.Positive
	case services.Stopped:
		return "square", theme.InkSoft
	default:
		return "circle-x", theme.Danger
	}
}

// lineColour tells streams apart: errors in red, the launcher's own notes in the primary colour.
func lineColour(context *ui.Context, stream domain.Stream) draw2d.Color {
	switch stream {
	case domain.Stderr:
		return context.Theme.Danger
	case domain.System:
		return context.Theme.Primary
	default:
		return context.Theme.Ink
	}
}

// runSummary is the one-line state of a run for a header.
func runSummary(snapshot services.RunSnapshot) string {
	duration := formatDuration(snapshot.Duration)
	switch snapshot.Status {
	case services.Running:
		return fmt.Sprintf("en marcha con %s · %s", snapshot.Runner, duration)
	case services.Succeeded:
		return fmt.Sprintf("código 0 · %s", duration)
	case services.Stopped:
		return fmt.Sprintf("parado · %s", duration)
	default:
		return fmt.Sprintf("código %d · %s", snapshot.ExitCode, duration)
	}
}

package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// Calibre3D drives the calibre3d project: choose its folder, set up its environment in one click,
// and save photos from its RealSense camera.
type Calibre3D struct {
	env      *services.PythonEnvService
	camera   *services.CaptureService
	preview  *services.PreviewService
	settings data.SettingsRepository
	project  string
	status   domain.EnvStatus
	reason   string        // why the environment is not ready, when the check could tell
	hasReqs  bool          // the project has a requirements.txt
	stale    bool          // the status has not been checked since the profile was built
	checked  *services.Run // the setup run whose end was already checked
	choosing bool          // the folder picker is shown
	naming   bool          // the photo name dialog is shown
	opening  bool          // the name dialog opens on this frame
	alias    string        // the photo name being typed
	capture  photoCapture  // the last photo started
	activity activity      // what the right panel shows
	view     components.CameraView
}

// activity is the last thing launched from the profile, shown on the right: a run with its output,
// or the live camera with its messages.
type activity struct {
	title string
	run   *services.Run
	live  bool
}

// NewCalibre3D builds the profile on the folder saved last time; its environment is checked when first shown.
func NewCalibre3D(env *services.PythonEnvService, camera *services.CaptureService, preview *services.PreviewService, settings data.SettingsRepository) (*Calibre3D, error) {
	saved, err := settings.Load()
	if err != nil {
		return nil, err
	}
	return &Calibre3D{env: env, camera: camera, preview: preview, settings: settings, project: saved.Calibre3DProject, stale: true}, nil
}

// ID returns the Calibre3D profile's ID.
func (*Calibre3D) ID() domain.ProfileID { return domain.Calibre3DProfileID }

// Name returns "Calibre3D".
func (*Calibre3D) Name() string { return "Calibre3D" }

// Draw shows the project, environment and camera cards on the left and the last activity on the right.
func (profile *Calibre3D) Draw(context *ui.Context, current *session.Session) {
	if profile.stale {
		profile.stale = false
		profile.check(current)
	}
	profile.checkAfterSetup(current)
	profile.announceCapture(current)
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		column := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(360), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
		context.Frame(column, ui.Look{Fill: context.Theme.Sunken}, func() {
			profile.drawProjectCard(context, current)
			profile.drawEnvCard(context, current)
			profile.drawCameraCard(context, current)
			context.Caption("Próximamente: plano, centrado y medidas.")
		})
		context.Separator()
		profile.drawActivity(context)
	})
	profile.drawFolderPicker(context, current)
	profile.drawNameDialog(context, current)
}

// Leave closes the live camera when another profile is chosen, so it is not left busy.
func (profile *Calibre3D) Leave(current *session.Session) {
	if err := profile.preview.Stop(); err != nil {
		current.Warn(err)
	}
}

// show makes a run the one the right panel follows.
func (profile *Calibre3D) show(title string, run *services.Run) {
	profile.activity = activity{title: title, run: run}
}

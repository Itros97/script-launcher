package profiles

import (
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views/components"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// Calibre3DServices are what the Calibre3D profile works with.
type Calibre3DServices struct {
	Env      *services.PythonEnvService
	Camera   *services.CaptureService
	Preview  *services.PreviewService
	Calibre  *services.CalibreService
	Settings data.SettingsRepository
}

// Calibre3D drives the calibre3d project: choose its folder, set up its environment in one click,
// see and photograph what its RealSense camera sees, and calibrate and measure pieces.
type Calibre3D struct {
	Calibre3DServices
	project  string
	status   domain.EnvStatus
	reason   string        // why the environment is not ready, when the check could tell
	hasReqs  bool          // the project has a requirements.txt
	stale    bool          // the status has not been checked since the profile was built
	checked  *services.Run // the setup run whose end was already checked
	choosing bool          // the folder picker is shown
	naming   nameKind      // what the name dialog asks for, if shown
	opening  bool          // the name dialog opens on this frame
	name     string        // the name being typed
	capture  photoCapture  // the last photo started
	step     stepRun       // the last calibre3d step started
	plane    planeState    // the project's reference plane
	result   *measured     // the last measurement shown
	activity activity      // what the right panel shows
	view     components.CameraView
	images   components.ImageFiles
}

// activity is the last thing launched from the profile, shown on the right: a run with its output
// (and a step's pictures), or the live camera with its messages.
type activity struct {
	title string
	run   *services.Run
	live  bool
	step  *stepRun
}

// NewCalibre3D builds the profile on the folder saved last time; its environment is checked when first shown.
func NewCalibre3D(with Calibre3DServices) (*Calibre3D, error) {
	saved, err := with.Settings.Load()
	if err != nil {
		return nil, err
	}
	return &Calibre3D{Calibre3DServices: with, project: saved.Calibre3DProject, stale: true}, nil
}

// ID returns the Calibre3D profile's ID.
func (*Calibre3D) ID() domain.ProfileID { return domain.Calibre3DProfileID }

// Name returns "Calibre3D".
func (*Calibre3D) Name() string { return "Calibre3D" }

// Draw shows the cards on the left, scrolling when they do not fit, and the last activity on the right.
func (profile *Calibre3D) Draw(context *ui.Context, current *session.Session) {
	if profile.stale {
		profile.stale = false
		profile.check(current)
	}
	profile.checkAfterSetup(current)
	profile.announceCapture(current)
	profile.finishStep(current)
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		context.Frame(ui.Layout{Width: ui.Fixed(360), Height: ui.Grow()}, ui.Look{Fill: context.Theme.Sunken}, func() {
			column := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
			context.ScrollArea("calibre3d-cards", column, func() {
				profile.drawProjectCard(context, current)
				profile.drawEnvCard(context, current)
				profile.drawMeasureCard(context, current)
				profile.drawCameraCard(context, current)
			})
		})
		context.Separator()
		profile.drawActivity(context)
	})
	profile.drawFolderPicker(context, current)
	profile.drawNameDialog(context, current)
}

// Leave closes the live camera when another profile is chosen, so it is not left busy.
func (profile *Calibre3D) Leave(current *session.Session) {
	if err := profile.Preview.Stop(); err != nil {
		current.Warn(err)
	}
}

// show makes a run the one the right panel follows.
func (profile *Calibre3D) show(title string, run *services.Run) {
	profile.activity = activity{title: title, run: run}
}

// busy tells whether something that uses the camera or the project is running now.
func (profile *Calibre3D) busy() bool {
	return profile.setupRunning() || profile.capturing() || profile.stepRunning()
}

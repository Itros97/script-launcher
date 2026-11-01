package views

import (
	"errors"
	"path/filepath"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// noticeDuration is how long a toast stays on screen.
const noticeDuration = 4 * time.Second

// AppView is the whole window: it holds the UI state and turns clicks into repository and run calls.
type AppView struct {
	repository data.ScriptRepository
	service    *services.RunService
	now        func() time.Time
	home       string
	scripts    []domain.Script
	selectedID string
	picking    bool
	notice     string
	warning    bool
	noticeEnd  time.Time
}

// NewAppView loads the saved scripts and builds the view; home is shown as ~ in paths. A corrupt list is an
// error so it is never overwritten.
func NewAppView(repository data.ScriptRepository, service *services.RunService, now func() time.Time, home string) (*AppView, error) {
	scripts, err := repository.List()
	if err != nil {
		return nil, err
	}
	app := &AppView{repository: repository, service: service, now: now, home: home, scripts: scripts}
	if len(scripts) > 0 {
		app.selectedID = scripts[0].ID
	}
	return app, nil
}

// Draw describes the window for this frame, top to bottom.
func (app *AppView) Draw(context *ui.Context) {
	context.KeyboardNavigation = ui.Always
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		app.drawToolbar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			app.drawScriptList(context)
			context.Separator()
			app.drawOutput(context)
		})
		context.Separator()
		app.drawStatusBar(context)
	})
	app.drawPicker(context)
	app.drawNotice(context)
}

// ---- actions ----

// addScript validates and saves a picked file, then selects it.
func (app *AppView) addScript(path string) {
	if _, err := app.service.Resolve(path); err != nil {
		app.notify(err.Error(), true)
		return
	}
	script, err := app.repository.Add(path)
	if errors.Is(err, data.ErrDuplicateScript) {
		app.selectedID = domain.ScriptID(path)
		app.notify("That script is already in the list", false)
		return
	}
	if err != nil {
		app.notify(err.Error(), true)
		return
	}
	app.scripts = append(app.scripts, script)
	app.selectedID = script.ID
	app.notify("Saved "+script.Name, false)
}

// removeScript stops and forgets a script.
func (app *AppView) removeScript(script domain.Script) {
	if err := app.service.Stop(script.ID); err != nil {
		app.notify(err.Error(), true)
		return
	}
	if err := app.repository.Remove(script.ID); err != nil {
		app.notify(err.Error(), true)
		return
	}
	scripts, err := app.repository.List()
	if err != nil {
		app.notify(err.Error(), true)
		return
	}
	app.scripts = scripts
	if app.selectedID == script.ID {
		app.selectedID = ""
	}
}

// runScript starts a script and selects it so its output is shown.
func (app *AppView) runScript(script domain.Script) {
	app.selectedID = script.ID
	if _, err := app.service.Start(script); err != nil {
		app.notify(err.Error(), true)
	}
}

// stopScript stops a running script.
func (app *AppView) stopScript(script domain.Script) {
	if err := app.service.Stop(script.ID); err != nil {
		app.notify(err.Error(), true)
	}
}

// ---- helpers ----

func (app *AppView) notify(text string, warning bool) {
	app.notice, app.warning, app.noticeEnd = text, warning, app.now().Add(noticeDuration)
}

func (app *AppView) selected() (domain.Script, bool) {
	for _, script := range app.scripts {
		if script.ID == app.selectedID {
			return script, true
		}
	}
	return domain.Script{}, false
}

// pickerStart opens the file picker where the last saved script lives.
func (app *AppView) pickerStart() string {
	if len(app.scripts) == 0 {
		return ""
	}
	return filepath.Dir(app.scripts[len(app.scripts)-1].Path)
}

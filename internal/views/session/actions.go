package session

import (
	"errors"
	"path/filepath"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
)

// ---- actions ----

// Select makes a script the selected one.
func (session *Session) Select(id string) { session.selectedID = id }

// Notify shows a message for a few seconds; warning marks it as a problem.
func (session *Session) Notify(text string, warning bool) {
	session.notice = Notice{Text: text, Warning: warning, Ends: session.now().Add(noticeDuration)}
}

// Add validates and saves a file, then selects it.
func (session *Session) Add(path string) {
	if _, err := session.service.Resolve(path); err != nil {
		session.Warn(err)
		return
	}
	script, err := session.repository.Add(path)
	if errors.Is(err, data.ErrDuplicateScript) {
		session.selectedID = domain.ScriptID(path)
		session.Notify("Ese script ya está en la lista", false)
		return
	}
	if err != nil {
		session.Warn(err)
		return
	}
	session.scripts = append(session.scripts, script)
	session.selectedID = script.ID
	session.Notify("Guardado: "+script.Name, false)
}

// Remove stops and forgets a script; the file on disk stays.
func (session *Session) Remove(script domain.Script) {
	if err := session.service.Stop(script.ID); err != nil {
		session.Warn(err)
		return
	}
	if err := session.repository.Remove(script.ID); err != nil {
		session.Warn(err)
		return
	}
	scripts, err := session.repository.List()
	if err != nil {
		session.Warn(err)
		return
	}
	session.scripts = scripts
	if session.selectedID == script.ID {
		session.selectedID = ""
	}
}

// Run starts a script and selects it so its output is shown.
func (session *Session) Run(script domain.Script) {
	session.selectedID = script.ID
	if _, err := session.service.Start(script); err != nil {
		session.Warn(err)
	}
}

// Stop stops a running script.
func (session *Session) Stop(script domain.Script) {
	if err := session.service.Stop(script.ID); err != nil {
		session.Warn(err)
	}
}

// ---- file picker ----

// OpenPicker asks for the file picker to be shown.
func (session *Session) OpenPicker() { session.picking = true }

// ClosePicker hides the file picker.
func (session *Session) ClosePicker() { session.picking = false }

// Picking tells whether the file picker is shown.
func (session *Session) Picking() bool { return session.picking }

// PickerStart is where the picker opens: the folder of the last saved script.
func (session *Session) PickerStart() string {
	if len(session.scripts) == 0 {
		return ""
	}
	return filepath.Dir(session.scripts[len(session.scripts)-1].Path)
}

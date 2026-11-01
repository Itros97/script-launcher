// Package session holds what every profile shares: the saved scripts, the selection, the runs and
// the notices, and the actions that change them.
package session

import (
	"time"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// noticeDuration is how long a notice stays on screen.
const noticeDuration = 4 * time.Second

// Session is the state shared by every profile's layout; it outlives a change of profile.
type Session struct {
	repository data.ScriptRepository
	service    *services.RunService
	now        func() time.Time
	home       string
	scripts    []domain.Script
	selectedID string
	picking    bool
	notice     Notice
}

// Notice is a short message for the user, shown until it expires.
type Notice struct {
	Text    string
	Warning bool
	Ends    time.Time
}

// New loads the saved scripts and selects the first one; home is shown as ~ in paths. A corrupt list is
// an error so it is never overwritten.
func New(repository data.ScriptRepository, service *services.RunService, now func() time.Time, home string) (*Session, error) {
	scripts, err := repository.List()
	if err != nil {
		return nil, err
	}
	session := &Session{repository: repository, service: service, now: now, home: home, scripts: scripts}
	if len(scripts) > 0 {
		session.selectedID = scripts[0].ID
	}
	return session, nil
}

// ---- reading ----

// Scripts returns the saved scripts in the order they were added.
func (session *Session) Scripts() []domain.Script { return session.scripts }

// Home returns the home folder, to shorten paths; empty shows them in full.
func (session *Session) Home() string { return session.home }

// Now returns the session's clock time.
func (session *Session) Now() time.Time { return session.now() }

// IsSelected tells whether this script is the selected one.
func (session *Session) IsSelected(id string) bool { return session.selectedID == id }

// Selected returns the selected script, if any.
func (session *Session) Selected() (domain.Script, bool) {
	for _, script := range session.scripts {
		if script.ID == session.selectedID {
			return script, true
		}
	}
	return domain.Script{}, false
}

// Latest returns the most recent run of a script in this session.
func (session *Session) Latest(id string) (*services.Run, bool) { return session.service.Latest(id) }

// RunningCount returns how many saved scripts are running now.
func (session *Session) RunningCount() int {
	count := 0
	for _, script := range session.scripts {
		if run, ok := session.service.Latest(script.ID); ok && run.Running() {
			count++
		}
	}
	return count
}

// CurrentNotice returns the notice still on screen, if any.
func (session *Session) CurrentNotice() (Notice, bool) {
	return session.notice, session.notice.Text != "" && session.now().Before(session.notice.Ends)
}

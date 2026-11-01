// Package profiles holds the window's layouts. A profile draws the body of the window from the
// shared session; the toolbar, the status bar and the file picker stay the same in every profile.
package profiles

import (
	"slices"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/views/session"
)

// Profile is one layout of the window's body.
type Profile interface {
	// ID is what the settings file stores.
	ID() domain.ProfileID
	// Name is shown in the profile selector.
	Name() string
	// Draw describes the body of the window for this frame.
	Draw(context *ui.Context, current *session.Session)
}

// Leaver is a profile that has something to close (a camera, a stream) when another profile is chosen.
type Leaver interface {
	// Leave is called once when the selector moves away from the profile.
	Leave(current *session.Session)
}

// Registry is the list of profiles the selector offers; the first one is the fallback.
type Registry struct {
	profiles []Profile
}

// NewRegistry builds a registry; it needs at least one profile, the fallback.
func NewRegistry(profiles ...Profile) Registry {
	if len(profiles) == 0 {
		panic("a profile registry needs at least one profile, the fallback")
	}
	return Registry{profiles: profiles}
}

// DefaultRegistry offers the general layout first, then Calibre3D.
func DefaultRegistry(calibre *Calibre3D) Registry { return NewRegistry(Default{}, calibre) }

// Profiles returns the profiles in the selector's order.
func (registry Registry) Profiles() []Profile { return registry.profiles }

// Names returns the profiles' names in the selector's order.
func (registry Registry) Names() []string {
	names := make([]string, len(registry.profiles))
	for index, profile := range registry.profiles {
		names[index] = profile.Name()
	}
	return names
}

// IndexOf returns where the profile with this ID is, or 0 (the fallback) when it is unknown.
func (registry Registry) IndexOf(id domain.ProfileID) int {
	return max(slices.IndexFunc(registry.profiles, func(profile Profile) bool { return profile.ID() == id }), 0)
}

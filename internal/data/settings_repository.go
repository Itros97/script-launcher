package data

import "github.com/Itros97/script-launcher/internal/domain"

// Settings are the user's choices that outlive the window.
type Settings struct {
	Profile domain.ProfileID `json:"profile"`
	// Calibre3DProject is the folder of the calibre3d project, empty until one is chosen.
	Calibre3DProject string `json:"calibre3dProject,omitempty"`
}

// DefaultSettings are used when nothing was saved yet.
func DefaultSettings() Settings { return Settings{Profile: domain.DefaultProfileID} }

// SettingsRepository keeps the settings between runs of the app.
type SettingsRepository interface {
	// Load returns the saved settings, or DefaultSettings when none were saved.
	Load() (Settings, error)
	// Save replaces the saved settings.
	Save(settings Settings) error
	// Update loads the settings, applies change and saves them, so each writer keeps the others' fields.
	Update(change func(settings *Settings)) error
}

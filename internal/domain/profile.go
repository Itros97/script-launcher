package domain

// ProfileID names a window layout; it is what the settings file stores.
type ProfileID string

const (
	DefaultProfileID   ProfileID = "default"   // the list of scripts beside their output
	Calibre3DProfileID ProfileID = "calibre3d" // the Calibre3D layout
)

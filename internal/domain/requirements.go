package domain

// Platform is a family of computers with its own requirements file, since some libraries are
// published under a different name on each.
type Platform string

const (
	Linux Platform = "linux" // Linux and any other computer that is not a Mac
	MacOS Platform = "macos"
)

// Platforms lists every platform with a requirements template, in the order they are offered.
var Platforms = []Platform{Linux, MacOS}

// PlatformFor returns the system a Go GOOS belongs to.
func PlatformFor(goos string) Platform {
	if goos == "darwin" {
		return MacOS
	}
	return Linux
}

// Name is how the system is called in the interface.
func (system Platform) Name() string {
	if system == MacOS {
		return "macOS"
	}
	return "Linux"
}

// RequirementsFile is the requirements file only this system installs, such as requirements-macos.txt.
func (system Platform) RequirementsFile() string { return "requirements-" + string(system) + ".txt" }

// RequirementsFor picks the requirements file setup installs on system: its own when the project
// has it, requirements.txt otherwise, and "" when it has neither.
func RequirementsFor(system Platform, has func(name string) bool) string {
	for _, name := range []string{system.RequirementsFile(), RequirementsFile} {
		if has(name) {
			return name
		}
	}
	return ""
}

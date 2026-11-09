package domain

import "testing"

func TestSystemForMacIsMacOSAndEverythingElseLinux(t *testing.T) {
	for goos, want := range map[string]Platform{"darwin": MacOS, "linux": Linux, "windows": Linux} {
		if got := PlatformFor(goos); got != want {
			t.Errorf("%s: got %s, want %s", goos, got, want)
		}
	}
}

func TestRequirementsForPrefersTheSystemsOwnFile(t *testing.T) {
	files := map[string]bool{"requirements.txt": true, "requirements-macos.txt": true, "requirements-linux.txt": true}
	has := func(name string) bool { return files[name] }
	if got := RequirementsFor(MacOS, has); got != "requirements-macos.txt" {
		t.Errorf("macOS: got %q", got)
	}
	if got := RequirementsFor(Linux, has); got != "requirements-linux.txt" {
		t.Errorf("Linux: got %q", got)
	}
}

func TestRequirementsForFallsBackToRequirementsTxt(t *testing.T) {
	files := map[string]bool{"requirements.txt": true, "requirements-linux.txt": true}
	if got := RequirementsFor(MacOS, func(name string) bool { return files[name] }); got != RequirementsFile {
		t.Errorf("got %q", got)
	}
	if got := RequirementsFor(MacOS, func(string) bool { return false }); got != "" {
		t.Errorf("no files: got %q", got)
	}
}

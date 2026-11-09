// Package templates embeds the files the launcher can hand to a project.
package templates

import (
	_ "embed"

	"github.com/Itros97/script-launcher/internal/domain"
)

var (
	//go:embed requirements-linux.txt
	linux []byte
	//go:embed requirements-macos.txt
	macOS []byte
)

// Requirements is the calibre3d requirements template for system.
func Requirements(system domain.Platform) []byte {
	if system == domain.MacOS {
		return macOS
	}
	return linux
}

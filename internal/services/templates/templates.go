// Package templates embeds the files the launcher can hand to a project.
package templates

import _ "embed"

// Requirements is the requirements.txt template for the calibre3d project.
//
//go:embed requirements.txt
var Requirements []byte

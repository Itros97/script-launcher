// Command screenshots draws the README's pictures: it saves a few sample scripts in a temporary
// folder, runs one to the end, and captures the window in the light and the dark theme.
//
//	go run ./cmd/screenshots -out docs/images
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/akrck02/go-ui/window"

	"github.com/Itros97/script-launcher/internal/data"
	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
	"github.com/Itros97/script-launcher/internal/views"
)

// samples are the scripts shown in the pictures; backup-notes.sh is run before capturing.
var samples = map[string]string{
	"backup-notes.sh":    "echo \"Backing up 3 folders\"\nfor f in docs photos music; do echo \"  copied $f\"; done\necho \"warning: photos/raw skipped (no space)\" >&2\necho \"Done\"\n",
	"clean-downloads.py": "print('Nothing older than 30 days')\n",
	"deploy":             "#!/bin/sh\necho deployed\n",
}

func main() {
	out := flag.String("out", filepath.Join("docs", "images"), "folder for hero-light.png and hero-dark.png")
	flag.Parse()
	if err := capture(*out); err != nil {
		fmt.Fprintln(os.Stderr, "screenshots:", err)
		os.Exit(1)
	}
}

// capture prepares the sample scripts and draws one picture per theme.
func capture(out string) error {
	folder, err := os.MkdirTemp("", "script-launcher-demo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(folder)
	repository := data.NewJSONScriptRepository(data.OSFileStore{}, filepath.Join(folder, "scripts.json"), time.Now)
	for _, name := range []string{"backup-notes.sh", "clean-downloads.py", "deploy"} {
		path := filepath.Join(folder, name)
		if err := os.WriteFile(path, []byte(samples[name]), 0o755); err != nil {
			return err
		}
		if _, err := repository.Add(path); err != nil {
			return err
		}
	}
	service := services.NewRunService(domain.DefaultRegistry(), time.Now, func() {})
	scripts, err := repository.List()
	if err != nil {
		return err
	}
	run, err := service.Start(scripts[0])
	if err != nil {
		return err
	}
	for run.Running() {
		time.Sleep(10 * time.Millisecond)
	}
	for _, dark := range []bool{false, true} {
		app, err := views.NewAppView(repository, service, time.Now, "")
		if err != nil {
			return err
		}
		name := map[bool]string{false: "hero-light.png", true: "hero-dark.png"}[dark]
		options := window.Options{Title: "Script Launcher", Width: 1100, Height: 560, Dark: dark, CapturePath: filepath.Join(out, name)}
		if err := window.Run(options, app.Draw); err != nil {
			return fmt.Errorf("capturing %s: %w", name, err)
		}
	}
	return nil
}

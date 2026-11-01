// Command script-launcher opens a window to save scripts and run them with one click.
//
//	go run ./cmd/script-launcher
//	go run ./cmd/script-launcher -dark
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

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "script-launcher:", err)
		os.Exit(1)
	}
}

// run wires the layers together and keeps the window open until it is closed.
func run() error {
	options := window.Options{Title: "Script Launcher", Width: 1100, Height: 680, MinWidth: 760, MinHeight: 420, Dark: window.SystemPrefersDark()}
	options.RegisterFlags(flag.CommandLine)
	configPath := flag.String("config", defaultConfigPath(), "file where the saved scripts are kept")
	flag.Parse()

	repository := data.NewJSONScriptRepository(data.OSFileStore{}, *configPath, time.Now)
	service := services.NewRunService(domain.DefaultRegistry(), time.Now, window.Wake)
	home, _ := os.UserHomeDir() // without a home folder, paths are simply shown in full
	app, err := views.NewAppView(repository, service, time.Now, home)
	if err != nil {
		return err
	}
	runErr := window.Run(options, app.Draw)
	if stopErr := service.StopAll(); stopErr != nil {
		return fmt.Errorf("stopping scripts on exit: %w", stopErr)
	}
	return runErr
}

// defaultConfigPath is scripts.json in the user's config folder ($XDG_CONFIG_HOME or ~/.config on Linux).
func defaultConfigPath() string {
	folder, err := os.UserConfigDir()
	if err != nil {
		folder = "."
	}
	return filepath.Join(folder, "script-launcher", "scripts.json")
}

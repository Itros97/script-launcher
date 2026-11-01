// Command script-launcher opens a window to save scripts and run them with one click.
//
//	go run ./cmd/script-launcher
//	go run ./cmd/script-launcher -dark
package main

import (
	"errors"
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
	"github.com/Itros97/script-launcher/internal/views/profiles"
	"github.com/Itros97/script-launcher/internal/views/session"
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
	configDir := flag.String("config-dir", defaultConfigDir(), "folder for scripts.json and settings.json")
	flag.Parse()

	scripts := data.NewJSONScriptRepository(data.OSFileStore{}, filepath.Join(*configDir, "scripts.json"), time.Now)
	settings := data.NewJSONSettingsRepository(data.OSFileStore{}, filepath.Join(*configDir, "settings.json"))
	service := services.NewRunService(domain.DefaultRegistry(), time.Now, window.Wake)
	home, _ := os.UserHomeDir() // without a home folder, paths are simply shown in full
	current, err := session.New(scripts, service, time.Now, home)
	if err != nil {
		return err
	}
	env := services.NewPythonEnvService(service, "python3", "calibre3d")
	preview := services.NewPreviewService(cacheDir(), window.Wake)
	camera := services.NewCaptureService(service, preview, cacheDir())
	calibre, err := profiles.NewCalibre3D(profiles.Calibre3DServices{
		Env: env, Camera: camera, Preview: preview, Calibre: services.NewCalibreService(service, preview), Settings: settings,
	})
	if err != nil {
		return err
	}
	app, err := views.NewAppView(current, profiles.DefaultRegistry(calibre), settings)
	if err != nil {
		return err
	}
	runErr := window.Run(options, app.Draw)
	if stopErr := errors.Join(service.StopAll(), preview.Stop()); stopErr != nil {
		return fmt.Errorf("stopping scripts on exit: %w", stopErr)
	}
	return runErr
}

// cacheDir is where helper scripts are written before they run ($XDG_CACHE_HOME or ~/.cache on Linux).
func cacheDir() string {
	folder, err := os.UserCacheDir()
	if err != nil {
		folder = os.TempDir()
	}
	return filepath.Join(folder, "script-launcher")
}

// defaultConfigDir is script-launcher in the user's config folder ($XDG_CONFIG_HOME or ~/.config on Linux).
func defaultConfigDir() string {
	folder, err := os.UserConfigDir()
	if err != nil {
		folder = "."
	}
	return filepath.Join(folder, "script-launcher")
}

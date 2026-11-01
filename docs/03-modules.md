# Modules

---

The code is split in layers, each in its own folder under `internal/`. Every layer only knows the
ones below it, and `cmd/script-launcher/main.go` is the only place that builds them and passes
each one what it needs.

```
views ─→ profiles ─→ components ─→ session ─→ services ─→ domain
                                       └────→ data ─────→ domain
```

| Layer | Files | What it holds |
|---|---|---|
| `domain` | `script.go`, `runner.go`, `output.go`, `profile.go` | Plain data and pure logic: a saved script, the runners and the registry that picks one, the output buffer, the profile IDs. No disk, no processes. |
| `data` | `script_repository.go`, `settings_repository.go`, their `json_*` implementations, `json_file.go` | The repositories and their JSON files. The disk sits behind a `FileStore`, so tests use a map. |
| `services` | `run_service.go`, `run.go`, `python_env_service.go`, `capture_service.go`, `preview_service.go` | Starting, watching and stopping processes; one `Run` per execution, safe to read from the UI while its goroutines write. Preparing and checking a Python `.venv`, saving photos, and the live camera view, with the scripts embedded in `capture/`. |
| `views/session` | `session.go`, `actions.go` | What every profile shares: the scripts, the selection, the notices, the file picker, and the actions (add, remove, run, stop). |
| `views/components` | `script_list.go`, `output_panel.go`, `bar.go`, … | The pieces layouts are built from. Each draws from a session and calls its actions. |
| `views/profiles` | `profile.go`, `default.go`, `calibre3d*.go` | The layouts, and the registry the selector offers. Calibre3D is split by card: project, environment, camera. |
| `views` | `app_view.go` | The window: toolbar with the profile selector, the active profile, status bar, picker, notices. |

## A frame

Go UI is immediate mode: there are no widget objects. Each frame `AppView.Draw` draws the toolbar,
then calls the active profile's `Draw` with the session, then the status bar. A run's goroutines
call `window.Wake()` after every line, so the window draws only when something changed and sleeps
otherwise.

## Adding a profile

A profile is anything that implements `profiles.Profile`:

```go
type Profile interface {
	ID() domain.ProfileID
	Name() string
	Draw(context *ui.Context, current *session.Session)
}
```

1. Add its ID to `internal/domain/profile.go`.
2. Write the layout in a new file in `internal/views/profiles/`, building it from `components`
   (`ScriptList`, `ScriptRow`, `RunButtons`, `OutputPanel`, `OutputLines`, `Bar`) and plain Go UI.
3. Add it to `DefaultRegistry` in `profile.go`. The selector shows profiles in that order; the
   first one is used when the saved profile is unknown.

A profile that needs state of its own (a chosen tab, a typed value) keeps it in its own struct
fields and is registered as a pointer. One that needs services gets them in a constructor, called
from `main.go`, as `NewCalibre3D(env, settings)` does.

## Adding a kind of script

A runner is anything that implements `domain.Runner`:

```go
type Runner interface {
	ID() string
	Matches(path string, executable bool) bool
	Command(path string) (program string, args []string)
}
```

For a new interpreter, add an `InterpreterRunner` to `DefaultRegistry` in
`internal/domain/runner.go`, before the `ExecutableRunner` fallback:

```go
InterpreterRunner{Program: "ruby", Extensions: []string{".rb"}},
```

Then add a case to `TestResolvesInterpreterByExtension`.

## Tests

```bash
go test -race ./...
```

- `domain` and `data` are tested with plain values and an in-memory `FileStore`.
- `services` runs the small scripts in `internal/services/testdata`, creates real `.venv`s, and
  saves photos through a stand-in `pyrealsense2` (`testdata/fake_realsense`) that returns a solid
  frame, checking the PNG it writes.
- `views` opens the window headless with Go UI's `uitest`, clicks the real buttons and the profile
  selector by name, and checks what the window shows.

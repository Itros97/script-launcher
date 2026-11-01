# Modules

---

The code is split in layers, each in its own folder under `internal/`. Every layer only knows the
ones below it, and `cmd/script-launcher/main.go` is the only place that builds them and passes
each one what it needs.

```
views     →  services  →  domain
   └────────→  data    →  domain
```

| Layer | Files | What it holds |
|---|---|---|
| `domain` | `script.go`, `runner.go`, `output.go` | Plain data and pure logic: a saved script, the runners and the registry that picks one, the output buffer. No disk, no processes. |
| `data` | `script_repository.go`, `json_script_repository.go` | The `ScriptRepository` interface and its JSON implementation. The disk sits behind a `FileStore`, so tests use a map. |
| `services` | `run_service.go`, `run.go` | Starting, watching and stopping processes. One `Run` per execution, safe to read from the UI while its goroutines write. |
| `views` | `app_view.go`, `chrome_view.go`, `script_list_view.go`, `output_view.go`, `format.go` | The window, drawn with Go UI. `AppView` holds the UI state and turns clicks into repository and service calls. |

## A frame

Go UI is immediate mode: there are no widget objects. Each frame `AppView.Draw` reads the list and
each script's latest `Run` snapshot and describes the window again. A run's goroutines call
`window.Wake()` after every line, so the window draws only when something changed and sleeps
otherwise.

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

## Adding a panel

Each panel is an `AppView` method in its own file (`drawScriptList`, `drawOutput`…). A new panel
is a new file with a `draw…` method, called from `Draw` in `app_view.go`.

## Tests

```bash
go test -race ./...
```

- `domain` and `data` are tested with plain values and an in-memory `FileStore`.
- `services` runs the small scripts in `internal/services/testdata`.
- `views` opens the window headless with Go UI's `uitest`, clicks the real buttons by name, and
  checks what the window shows.

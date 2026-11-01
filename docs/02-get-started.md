# Get started

---

## Build and run

You need Go 1.25 or newer, a C compiler for cgo, and a display with OpenGL 4.1.

```bash
go run ./cmd/script-launcher        # follows the system's light or dark theme
go run ./cmd/script-launcher -dark  # forces the dark theme
```

To keep it at hand, build a binary and put it in your `PATH`:

```bash
go build -o ~/.local/bin/script-launcher ./cmd/script-launcher
```

## Add and run a script

1. Press **Añadir script**. The picker opens in the folder of the last script you added.
2. Choose a file. It is saved and selected.
3. Press **▶** on its row. The output appears on the right and follows new lines until you
   scroll up.

The script runs in its own folder, so relative paths inside it work as they do from a terminal
opened there.

## Which program runs a file

| File | Runs with |
|---|---|
| `.sh`, `.bash` | `bash` |
| `.py` | `python3` |
| `.js`, `.mjs`, `.ts` | `node` |
| anything else with the executable bit | the file itself (its `#!` line decides) |

A file that matches none of these is refused when you add it.

## Where the list is kept

In `$XDG_CONFIG_HOME/script-launcher/`, which is `~/.config/script-launcher/` on most systems:

| File | Holds |
|---|---|
| `scripts.json` | The saved scripts. |
| `settings.json` | The profile used last. |

`-config-dir <folder>` uses another folder.

If either file is not valid JSON, the launcher refuses to start and says where it is, instead of
overwriting it. Fix the file or delete it.

# Overview

---

Script Launcher is a desktop window that keeps a list of your scripts and runs them with one click.
You find a script on disk once; after that it is a row in the list, ready to run.

## What it does

- **Add script** opens a file picker. The chosen file is checked (it must exist and be something
  the launcher knows how to run) and saved to the list.
- **▶** runs the script. Its standard output and error appear live on the right, errors in red.
- **■** stops it: the launcher sends `SIGTERM` to the script and every process it started.
- When the script ends, the last line says how: its exit code and how long it took. The row shows
  a green tick, a red cross or a grey square.
- **🗑** removes a script from the list. The file on disk is left alone.
- A script whose file was moved or deleted shows a warning sign and cannot be run.

## What it does not do yet

- Pass arguments or environment variables to a script.
- Read from the keyboard: a script's standard input is empty, so a script that asks a question
  gets no answer.
- Keep the output of past runs after the window closes.

See the [roadmap](04-roadmap.md).

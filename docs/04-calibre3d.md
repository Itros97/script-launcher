# Calibre3D

---

The Calibre3D profile is for the calibre3d project: a Python program that measures pieces with a
RealSense camera. It is meant for people who have never used a terminal: choose the project's
folder once, press one button, and everything it needs is installed.

## Using it

1. Choose **Calibre3D** in the profile selector.
2. Press **Elegir carpeta** and choose the folder of the calibre3d project (the one that holds the
   `calibre3d` package and `prueba.sh`). It is remembered.
3. Press **Preparar entorno**. The installation shows on the right, line by line.
4. When it ends, the status turns **Listo**.

## No requirements.txt?

The project does not need one. When it has none, the **Entorno** card says "Sin requirements.txt:
se usará la plantilla", and **Preparar entorno** first writes the launcher's template into the
project as `requirements.txt` and then installs it. The template works on every computer the
launcher ships for: each camera line carries a condition, so pip installs only the one for that
computer.

| Package | Installed on | Why |
|---|---|---|
| `numpy>=1.24` | all | Point clouds and depth images. |
| `scipy>=1.10` | all | Fitting the plane, segmenting and measuring the piece. |
| `matplotlib>=3.7` | all | The thickness heat maps and the views (PNG). |
| `pyrealsense2>=2.55` | Linux | The RealSense camera; **Ver cámara** and **Guardar imagen** need it too. |
| `pyrealsense2-macosx>=2.56` | Mac with Apple silicon | The same camera library. Intel publishes no macOS wheels; this community build imports the same way (`import pyrealsense2`) and needs macOS 15 or later. |

On an Intel Mac there is no camera library, so nothing is installed for it and the rest still works
(**Demo sin cámara** included). The conditions use only `sys_platform` and `platform_machine`:
`platform_release` would make pip fail on Linux, where it is a kernel name such as `6.8.0-45-generic`
and not a version.

A requirements file that lists no packages (only comments or blank lines) installs nothing and
skips pip, so it needs no internet. **Descargar plantilla de requirements.txt** writes the
template without installing; it never overwrites a file: if `requirements.txt` already exists, the
template is saved beside it as `requirements.plantilla.txt` to compare.

## The status

| Status | Means | What to do |
|---|---|---|
| 🔴 Sin proyecto | No folder chosen, or it is gone. | Press **Elegir carpeta**. |
| 🔴 Sin entorno | The folder has no `.venv`. | Press **Preparar entorno**. |
| 🟠 Incompleto | `.venv` exists, but `import calibre3d` fails in it. | Read the reason under the status (below). |

When the environment is incomplete, the card says why, from what Python answered:

| Reason shown | Means |
|---|---|
| En esta carpeta no está el programa calibre3d | The folder chosen is not the one with the `calibre3d` package: choose the project's folder, the one with `prueba.sh`. |
| Esta carpeta tiene prueba.sh pero le falta la carpeta calibre3d/ | The right folder, but only `prueba.sh` was copied: copy the whole project, `calibre3d/` included. |
| Falta la librería X: añade Y a requirements.txt | calibre3d imports a library that is not installed; Y is its pip name (`cv2` → `opencv-python`). |
| Error al cargar calibre3d: … | Another error while importing; the text is Python's last line. |
| 🟢 Listo | `import calibre3d` works in `.venv`. | Nothing. |

The status is checked when the profile is first shown, when a folder is chosen, when an
installation ends, and when you press **Comprobar**. It is not checked on every frame: the check
starts Python once, which takes a moment.

## What "Preparar entorno" does

In the project folder, one after the other, stopping at the first that fails:

```bash
python3 -m venv .venv
.venv/bin/python -m pip install --upgrade pip
.venv/bin/python -m pip install -r requirements.txt
```

Without a `requirements.txt` it writes the template first (see above); with one that lists no
packages it only creates `.venv` and says there was nothing to install. An
existing `.venv` is reused, not deleted. Pressing it again on a ready environment (**Volver a
preparar**) installs whatever is missing.

It needs an internet connection. The `python3` it starts from is, in this order: one the launcher
downloaded earlier, the computer's own (on macOS the python.org or Homebrew one; Apple's
`/usr/bin/python3` does not count, since without the developer tools it only opens a dialog), or,
when there is none, a fresh download. That download is a ready-made Python 3.12 from
[python-build-standalone](https://github.com/astral-sh/python-build-standalone) for Linux (x86-64,
ARM) and macOS (Apple silicon, Intel): no installer and no password. Its checksum is pinned in
`internal/domain/python_download.go`, so a damaged or altered file is refused. It is unpacked in
`python-3.12.15/` inside the config folder (`~/.config/script-launcher/` on Linux,
`~/Library/Application Support/script-launcher/` on macOS) and used from then on by every project.
Delete that folder to go back to the computer's own Python.

## Calibrating and measuring

Once the status is **Listo**, the **Medir** card runs calibre3d's steps in order, each with what to
do before pressing it:

| Button | Runs | Do first |
|---|---|---|
| **1. Plano** | `python -m calibre3d plano` | Leave the mat empty. Once done, it says when and how noisy ("Hecho el 1 nov, 22:10 · ruido 0,04 mm"), from `datos/plano.json`. |
| **2. Centrar** | `python -m calibre3d centrar` | Put a coin on the cross. Its output says how far to move the mat; `centrado_vista.png` shows under it. |
| **3. Medir** | `python -m calibre3d medir <name>` | Put the piece inside the circle. It asks for the piece's name (letters, digits, `-`, `_`); a name already measured is refused, so no result is overwritten. Without a plane it says "Haz primero el plano", but lets you go on, as calibre3d does. |
| **Demo sin cámara** | `python -m calibre3d demo` | Nothing: it measures a simulated piece, to try everything without the camera. |

While a step runs, its output shows live in **Actividad** and the other steps, the photo and the
camera wait. The live camera is closed before each step (with a notice), since calibre3d needs the
camera alone.

When **Medir** or the demo ends well, the launcher reads the result file calibre3d left
(`datos/resultados/<name>.json`, or `demo/resultados/demo_canto.json`) and shows a strip at the top
of **Actividad** with the length, width, thickness and maximum Feret diameter, and calibre3d's
warnings. It stays there, whatever runs next, until another measurement replaces it. Under the
output come the heat map (`_mapa.png`) and the view (`_vista.png`).

If a step fails, its error shows in red under **Actividad** and the last result stays as it was.

`patron` (calibrating with pieces of known size) and `plantilla` (the mat's PDF) are not buttons
yet: run them from a terminal in the project folder.

## Seeing the camera

Once the status is **Listo**, **Ver cámara** in the **Cámara** card shows the camera live in the
**Actividad** panel: the colour picture at 1280×720, about 15 frames a second, fitted to the
panel in 16:9, with the camera's messages under it. **Parar cámara** closes it.

The camera also closes by itself when you switch to another profile and when the window closes,
so it is never left busy. While it is open the window redraws with every frame, which costs some
processor; closed, the window goes back to sleeping until something happens.

Behind it, `preview.py` (shipped inside the launcher, run with the project's `.venv/bin/python`)
writes each frame to its standard output as `RSF1`, the width and height as big-endian 16-bit
numbers, and the RGB bytes. The launcher keeps only the latest frame and paints it.

## Saving a photo

Once the status is **Listo**, the **Cámara** card offers **Guardar imagen**:

1. Press **Guardar imagen**. A dialog asks for the photo's name.
2. Type a name (letters, digits, `-` and `_`, no spaces, up to 64 characters) and press **Guardar**
   or Enter. A name already used is refused, so nothing is overwritten.
3. With the camera open (**Ver cámara**), the photo is exactly the frame on screen, saved by the
   launcher without touching the camera. With it closed, the camera opens, skips 30 frames so its
   exposure settles, saves one colour frame at 1280×720 to `datos/fotos/<name>.png` inside the
   project, and closes again.

The capture is a small Python script that ships inside the launcher (`internal/services/capture/capture.py`).
It runs with the project's own `.venv/bin/python` and only needs `pyrealsense2`, which
`requirements.txt` installs; the PNG is written with Python's standard library. It does not depend
on how the calibre3d package is written.

Only one program can use the camera at a time: do not save a photo while another program has the
camera open.

| Message | Means |
|---|---|
| No se encuentra ninguna RealSense conectada. | No camera is connected, or it is not detected. |
| No está instalado pyrealsense2 | The environment lacks the camera library: press **Preparar entorno**. |
| La cámara no responde | The camera did not send frames for 5 seconds. |
| La cámara aún está arrancando | The live view has no frame yet: wait a moment and save again. |

## Limits

**Listo** means the package imports. It does not prove the camera is connected or that every
dependency imported only later (inside a command) is installed.

The photo capture, the live view and the steps that use the camera have only been tested against
stand-ins for `pyrealsense2` and calibre3d, not a real camera. The demo was run with the real
calibre3d project.

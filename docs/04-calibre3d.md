# Calibre3D

---

The Calibre3D profile is for the calibre3d project: a Python program that measures pieces with a
RealSense camera. It is meant for people who have never used a terminal: choose the project's
folder once, press one button, and everything it needs is installed.

## Using it

1. Choose **Calibre3D** in the profile selector.
2. Press **Elegir carpeta** and choose the folder of the calibre3d project (the one that holds the
   `calibre3d` package and `requirements.txt`). It is remembered.
3. Press **Preparar entorno**. The installation shows on the right, line by line.
4. When it ends, the status turns **Listo**.

## No requirements.txt?

If the project has no `requirements.txt`, the **Entorno** card says so. **Descargar plantilla de
requirements.txt** writes one into the project folder with what calibre3d needs:

| Package | Why |
|---|---|
| `pyrealsense2` | The RealSense camera; **Guardar imagen** needs it too. |
| `numpy` | Working with the colour and depth images. |
| `opencv-python` | Processing and saving pictures (the views and maps in `datos/resultados`). |

`numpy` and `opencv-python` are a guess at what calibre3d uses: add or remove lines to match the
project. The button never overwrites a file: if `requirements.txt` already exists, the template is
saved beside it as `requirements.plantilla.txt` to compare.

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

Without a `requirements.txt` it only creates `.venv` and says there was nothing to install. An
existing `.venv` is reused, not deleted. Pressing it again on a ready environment (**Volver a
preparar**) installs whatever is missing.

It needs `python3` with the `venv` module and, for the dependencies, an internet connection.

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

The calibration steps (`plano`, `centrar`, `medir`) are not buttons yet; see the
[roadmap](05-roadmap.md). The photo capture and the live view have only been tested against a
stand-in for `pyrealsense2`, not a real camera. When the calibration steps arrive, they will close
the live view before running calibre3d, since only one program can use the camera.

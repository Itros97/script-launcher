"""Stream colour frames from the first RealSense camera to standard output, until stopped.

Usage: python preview.py

Each frame is b"RSF1", the width and height as big-endian uint16, then width × height × 3 bytes of
RGB. Messages for the user go to standard error. SIGTERM stops it cleanly, releasing the camera.
"""

import signal
import struct
import sys

from camera import HEIGHT, WIDTH, CameraError, next_rgb, open_camera

FPS = 15
HEADER = b"RSF1" + struct.pack(">HH", WIDTH, HEIGHT)


def stop(*_):
	"""Turn SIGTERM into a normal exit, so the camera is released in the finally below."""
	sys.exit(0)


def preview() -> int:
	"""Send frames until the launcher stops reading or stops the process; returns the exit code."""
	signal.signal(signal.SIGTERM, stop)
	try:
		pipeline = open_camera(FPS)
	except CameraError as error:
		print(error, file=sys.stderr)
		return error.code
	print(f"▸ Cámara en marcha ({WIDTH}×{HEIGHT}, {FPS} imágenes por segundo)", file=sys.stderr, flush=True)
	out = sys.stdout.buffer
	try:
		while True:
			out.write(HEADER + next_rgb(pipeline))
			out.flush()
	except CameraError as error:
		print(error, file=sys.stderr)
		return error.code
	except BrokenPipeError:
		return 0
	finally:
		pipeline.stop()


if __name__ == "__main__":
	sys.exit(preview())

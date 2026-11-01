"""Save one colour frame from the first RealSense camera as a PNG.

Usage: python capture.py <target.png>

Only needs pyrealsense2 and the standard library: the PNG is written by hand with zlib.
"""

import struct
import sys
import zlib
from pathlib import Path

from camera import HEIGHT, WIDTH, CameraError, next_rgb, open_camera

FPS = 30
WARMUP_FRAMES = 30  # lets auto exposure settle before the frame that is kept


def png_chunk(kind: bytes, data: bytes) -> bytes:
	"""One PNG chunk: length, type, data and CRC."""
	return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)


def write_png(path: Path, width: int, height: int, rgb: bytes) -> None:
	"""Write tightly packed 8-bit RGB rows as a PNG file."""
	row = width * 3
	raw = b"".join(b"\x00" + rgb[y * row : (y + 1) * row] for y in range(height))
	header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
	png = b"\x89PNG\r\n\x1a\n" + png_chunk(b"IHDR", header) + png_chunk(b"IDAT", zlib.compress(raw, 6)) + png_chunk(b"IEND", b"")
	path.write_bytes(png)


def capture(target: Path) -> int:
	"""Grab one colour frame and save it; returns the process exit code."""
	print(f"▸ Abriendo la cámara ({WIDTH}×{HEIGHT})", flush=True)
	try:
		pipeline = open_camera(FPS)
	except CameraError as error:
		print(error, file=sys.stderr)
		return error.code
	try:
		for _ in range(WARMUP_FRAMES):
			next_rgb(pipeline)
		rgb = next_rgb(pipeline)
	except CameraError as error:
		print(error, file=sys.stderr)
		return error.code
	finally:
		pipeline.stop()
	target.parent.mkdir(parents=True, exist_ok=True)
	write_png(target, WIDTH, HEIGHT, rgb)
	print(f"✓ Imagen guardada en {target}", flush=True)
	return 0


if __name__ == "__main__":
	if len(sys.argv) != 2:
		print("Uso: capture.py <imagen.png>", file=sys.stderr)
		sys.exit(2)
	sys.exit(capture(Path(sys.argv[1])))

"""Fake calibre3d command line: plano, centrar, medir <nombre> and demo, writing files shaped like the real ones.

FAKE_CALIBRE_FAIL=1 makes every command fail as a missing camera would.
"""

import json
import os
import struct
import sys
import zlib
from datetime import datetime
from pathlib import Path


def png(path: Path, rgb=(30, 120, 200), width=8, height=4) -> None:
	"""Write a small solid PNG."""
	def chunk(kind, data):
		return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)
	raw = b"".join(b"\x00" + bytes(rgb) * width for _ in range(height))
	header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
	path.parent.mkdir(parents=True, exist_ok=True)
	path.write_bytes(b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(b"IDAT", zlib.compress(raw)) + chunk(b"IEND", b""))


def now() -> str:
	return datetime.now().isoformat(timespec="seconds")


def plano(folder: Path) -> None:
	folder.mkdir(parents=True, exist_ok=True)
	info = {"normal": [0, 0, -1], "d": 133.0, "altura_camara_mm": 133.0, "rms_mm": 0.04, "cobertura": 1.0, "fecha": now()}
	(folder / "plano.json").write_text(json.dumps(info))
	print("Plano guardado: cámara a 133.0 mm, ruido 0.040 mm RMS, 100% de la imagen en el plano.")


def medir(folder: Path, nombre: str) -> None:
	resultados = folder / "resultados"
	resultado = {"largo_mm": 61.234, "ancho_mm": 40.5, "espesor_mm": 19.875, "feret_max_mm": 62.0, "aviso": ["Pieza lejos del centro"]}
	datos = {"nombre": nombre, "fecha": now(), "resultado": resultado, "bruto": resultado, "correccion": {}}
	resultados.mkdir(parents=True, exist_ok=True)
	(resultados / f"{nombre}.json").write_text(json.dumps(datos, ensure_ascii=False))
	png(resultados / f"{nombre}_mapa.png", (200, 60, 30))
	png(resultados / f"{nombre}_vista.png")
	print("  Largo   :   61.234 mm")
	print(f"Resultados en {resultados}/{nombre}.*")


def main(argv) -> int:
	if os.environ.get("FAKE_CALIBRE_FAIL") == "1":
		print("RuntimeError: No device connected", file=sys.stderr)
		return 1
	command = argv[0]
	if command == "plano":
		plano(Path("datos"))
	elif command == "centrar":
		png(Path("datos") / "resultados" / "centrado_vista.png")
		print("Centrado (dentro de 2.0 mm). Fija la plantilla.")
	elif command == "medir":
		medir(Path("datos"), argv[1])
	elif command == "demo":
		plano(Path("demo"))
		medir(Path("demo"), "demo_canto")
	else:
		print(f"orden desconocida: {command}", file=sys.stderr)
		return 2
	return 0


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))

"""Opening the RealSense camera, shared by capture.py and preview.py.

Messages are in Spanish because the launcher shows them to the user as they are.
"""

WIDTH, HEIGHT = 1280, 720
FRAME_TIMEOUT_MS = 5000


class CameraError(Exception):
	"""A problem the user can act on, with the exit code that reports it."""

	def __init__(self, message: str, code: int):
		super().__init__(message)
		self.code = code


def open_camera(fps: int):
	"""Start the colour stream of the first camera, or raise CameraError saying why not."""
	try:
		import pyrealsense2 as rs
	except ImportError:
		raise CameraError("No está instalado pyrealsense2: pulsa «Preparar entorno».", 3)
	if len(rs.context().query_devices()) == 0:
		raise CameraError("No se encuentra ninguna RealSense conectada.", 4)
	config = rs.config()
	config.enable_stream(rs.stream.color, WIDTH, HEIGHT, rs.format.rgb8, fps)
	pipeline = rs.pipeline()
	pipeline.start(config)
	return pipeline


def next_rgb(pipeline) -> bytes:
	"""Wait for the next colour frame and return its pixels as tightly packed RGB rows."""
	try:
		frame = pipeline.wait_for_frames(FRAME_TIMEOUT_MS).get_color_frame()
	except RuntimeError as error:
		raise CameraError(f"La cámara no responde: {error}", 5)
	if not frame:
		raise CameraError("La cámara no ha devuelto imagen a color.", 5)
	data, stride, row = bytes(frame.get_data()), frame.get_stride_in_bytes(), WIDTH * 3
	if stride == row:
		return data
	return b"".join(data[y * stride : y * stride + row] for y in range(HEIGHT))

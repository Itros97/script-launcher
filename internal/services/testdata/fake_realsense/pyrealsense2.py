"""A stand-in for pyrealsense2 in tests: one camera (none if FAKE_RS_DEVICES=0) that returns a solid orange frame."""

import os
import time

WIDTH, HEIGHT = 1280, 720


class stream:
	color = "color"


class format:
	rgb8 = "rgb8"


class context:
	def query_devices(self):
		return ["camera"] * int(os.environ.get("FAKE_RS_DEVICES", "1"))


class config:
	def enable_stream(self, *args):
		pass


class _Frame:
	def __bool__(self):
		return True

	def get_data(self):
		return bytes([255, 128, 0]) * (WIDTH * HEIGHT)

	def get_stride_in_bytes(self):
		return WIDTH * 3


class _Frames:
	def get_color_frame(self):
		return _Frame()


class pipeline:
	def start(self, config):
		pass

	def wait_for_frames(self, timeout_ms=5000):
		time.sleep(0.01)  # a real camera paces its frames; this keeps streaming tests off the CPU
		return _Frames()

	def stop(self):
		pass

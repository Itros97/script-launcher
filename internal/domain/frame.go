package domain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// FrameMagic opens every frame the preview script writes, so a stream that lost its place is noticed.
const FrameMagic = "RSF1"

// frameHeaderSize is the magic, then the width and the height as big-endian uint16.
const frameHeaderSize = len(FrameMagic) + 4

// maxFrameSide bounds a frame's width and height, so a corrupt header cannot ask for gigabytes.
const maxFrameSide = 4096

// ErrBadFrame is returned for a frame stream that does not follow the format.
var ErrBadFrame = errors.New("bad camera frame")

// Frame is one colour picture from the camera: Width × Height pixels, 3 bytes (R, G, B) each, row after row.
type Frame struct {
	Width, Height int
	RGB           []byte
}

// FrameHeader returns the bytes that go before a frame of this size.
func FrameHeader(width, height int) []byte {
	header := make([]byte, frameHeaderSize)
	copy(header, FrameMagic)
	binary.BigEndian.PutUint16(header[4:], uint16(width))
	binary.BigEndian.PutUint16(header[6:], uint16(height))
	return header
}

// ReadFrame reads the next frame from a stream into buffer, growing it when the frame is bigger,
// and returns the frame, whose pixels live in the returned buffer.
func ReadFrame(stream io.Reader, buffer []byte) (Frame, []byte, error) {
	header := make([]byte, frameHeaderSize)
	if _, err := io.ReadFull(stream, header); err != nil {
		return Frame{}, buffer, err
	}
	if string(header[:4]) != FrameMagic {
		return Frame{}, buffer, fmt.Errorf("%w: header starts with %q", ErrBadFrame, header[:4])
	}
	width, height := int(binary.BigEndian.Uint16(header[4:])), int(binary.BigEndian.Uint16(header[6:]))
	if width == 0 || height == 0 || width > maxFrameSide || height > maxFrameSide {
		return Frame{}, buffer, fmt.Errorf("%w: size %d×%d", ErrBadFrame, width, height)
	}
	size := width * height * 3
	if cap(buffer) < size {
		buffer = make([]byte, size)
	}
	buffer = buffer[:size]
	if _, err := io.ReadFull(stream, buffer); err != nil {
		return Frame{}, buffer, fmt.Errorf("%w: cut short: %v", ErrBadFrame, err)
	}
	return Frame{Width: width, Height: height, RGB: buffer}, buffer, nil
}

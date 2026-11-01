package domain

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestFramesRoundTripThroughAStream(t *testing.T) {
	var stream bytes.Buffer
	first, second := bytes.Repeat([]byte{1, 2, 3}, 4*2), bytes.Repeat([]byte{9, 8, 7}, 2*2)
	stream.Write(FrameHeader(4, 2))
	stream.Write(first)
	stream.Write(FrameHeader(2, 2))
	stream.Write(second)
	frame, buffer, err := ReadFrame(&stream, nil)
	if err != nil || frame.Width != 4 || frame.Height != 2 || !bytes.Equal(frame.RGB, first) {
		t.Fatalf("first frame: got %+v, %v", frame, err)
	}
	frame, _, err = ReadFrame(&stream, buffer)
	if err != nil || frame.Width != 2 || !bytes.Equal(frame.RGB, second) {
		t.Fatalf("second frame: got %+v, %v", frame, err)
	}
	if _, _, err := ReadFrame(&stream, buffer); !errors.Is(err, io.EOF) {
		t.Errorf("after the last frame: got %v, want EOF", err)
	}
}

func TestAStreamThatLostItsPlaceIsABadFrame(t *testing.T) {
	_, _, err := ReadFrame(bytes.NewReader([]byte("JUNKJUNK")), nil)
	if !errors.Is(err, ErrBadFrame) {
		t.Errorf("got %v, want ErrBadFrame", err)
	}
}

func TestAnAbsurdSizeIsABadFrame(t *testing.T) {
	_, _, err := ReadFrame(bytes.NewReader(FrameHeader(60000, 60000)), nil)
	if !errors.Is(err, ErrBadFrame) {
		t.Errorf("got %v, want ErrBadFrame", err)
	}
}

func TestACutFrameIsABadFrame(t *testing.T) {
	stream := append(FrameHeader(4, 4), 1, 2, 3)
	if _, _, err := ReadFrame(bytes.NewReader(stream), nil); !errors.Is(err, ErrBadFrame) {
		t.Errorf("got %v, want ErrBadFrame", err)
	}
}

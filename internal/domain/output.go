package domain

// Stream says where a line of output came from.
type Stream int

const (
	Stdout Stream = iota // the script's standard output
	Stderr               // the script's standard error
	System               // a note from the launcher itself (started, exit code)
)

// Line is one line of a run's output.
type Line struct {
	Text   string
	Stream Stream
}

// OutputBuffer keeps the last lines of a run's output, dropping the oldest past its limit.
type OutputBuffer struct {
	lines   []Line
	limit   int
	dropped int
}

// NewOutputBuffer builds a buffer that keeps at most limit lines.
func NewOutputBuffer(limit int) *OutputBuffer {
	if limit <= 0 {
		panic("output buffer limit must be positive")
	}
	return &OutputBuffer{limit: limit}
}

// Append adds a line, dropping the oldest one when the buffer is full.
func (buffer *OutputBuffer) Append(line Line) {
	buffer.lines = append(buffer.lines, line)
	overflow := len(buffer.lines) - buffer.limit
	if overflow <= 0 {
		return
	}
	buffer.lines = append(buffer.lines[:0], buffer.lines[overflow:]...)
	buffer.dropped += overflow
}

// Lines returns a copy of the lines kept, oldest first.
func (buffer *OutputBuffer) Lines() []Line {
	return append([]Line(nil), buffer.lines...)
}

// Dropped returns how many lines were discarded to stay within the limit.
func (buffer *OutputBuffer) Dropped() int { return buffer.dropped }

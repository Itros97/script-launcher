package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrInvalidSize is returned for a pattern's real size that is not a number above zero.
var ErrInvalidSize = errors.New("invalid size")

// Pattern is a piece of known size measured to calibrate calibre3d, with its real sizes in millimetres.
type Pattern struct {
	Name      string
	Length    float64
	Width     float64
	Thickness float64
}

// Correction is calibre3d's linear correction, real = K·measured + Offset, for length and width (XY) and thickness (Z).
type Correction struct {
	KXY      float64 `json:"k_xy"`
	OffsetXY float64 `json:"off_xy"`
	KZ       float64 `json:"k_z"`
	OffsetZ  float64 `json:"off_z"`
}

// Calibration is what calibre3d's calibracion.json holds: how many patterns were measured and the correction they give.
type Calibration struct {
	Patterns   int
	Correction Correction
}

// calibrationFile is the part of calibracion.json the launcher shows.
type calibrationFile struct {
	Patterns   []json.RawMessage `json:"patrones"`
	Correction *Correction       `json:"correccion"`
}

// ParsePattern checks a pattern typed in the launcher: a valid name and three real sizes above zero, with a decimal point or comma.
func ParsePattern(name, length, width, thickness string) (Pattern, error) {
	if err := ValidatePhotoAlias(name); err != nil {
		return Pattern{}, err
	}
	sizes := make([]float64, 3)
	for index, text := range []string{length, width, thickness} {
		size, err := parseSize(text)
		if err != nil {
			return Pattern{}, err
		}
		sizes[index] = size
	}
	return Pattern{Name: name, Length: sizes[0], Width: sizes[1], Thickness: sizes[2]}, nil
}

// parseSize reads a size in millimetres above zero, with a decimal point or comma.
func parseSize(text string) (float64, error) {
	size, err := strconv.ParseFloat(strings.Replace(strings.TrimSpace(text), ",", ".", 1), 64)
	if err != nil || size <= 0 {
		return 0, fmt.Errorf("%w: %q", ErrInvalidSize, text)
	}
	return size, nil
}

// Args are the arguments that measure the pattern with `python` and add it to the calibration.
func (pattern Pattern) Args() []string {
	return append(StepPattern.Args(pattern.Name), "--real", formatSize(pattern.Length), formatSize(pattern.Width), formatSize(pattern.Thickness))
}

// formatSize writes a size as Python reads it, with no more digits than it has.
func formatSize(size float64) string { return strconv.FormatFloat(size, 'f', -1, 64) }

// ParseCalibration reads calibre3d's calibracion.json.
func ParseCalibration(content []byte) (Calibration, error) {
	var file calibrationFile
	if err := json.Unmarshal(content, &file); err != nil {
		return Calibration{}, fmt.Errorf("reading the calibration: %w", err)
	}
	if file.Correction == nil {
		return Calibration{}, fmt.Errorf("reading the calibration: it has no \"correccion\"")
	}
	return Calibration{Patterns: len(file.Patterns), Correction: *file.Correction}, nil
}

// CalibrationFile is where calibre3d keeps the patterns measured and their correction.
func CalibrationFile(project string) string {
	return filepath.Join(project, calibreDataDir, "calibracion.json")
}

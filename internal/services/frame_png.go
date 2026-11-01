package services

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/Itros97/script-launcher/internal/domain"
)

// SaveFramePNG writes a camera frame to path as a PNG, creating its folder; it never overwrites a file.
func SaveFramePNG(frame domain.Frame, path string) error {
	picture := image.NewNRGBA(image.Rect(0, 0, frame.Width, frame.Height))
	for source, target := 0, 0; source+2 < len(frame.RGB); source, target = source+3, target+4 {
		picture.Pix[target], picture.Pix[target+1], picture.Pix[target+2], picture.Pix[target+3] = frame.RGB[source], frame.RGB[source+1], frame.RGB[source+2], 255
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("saving %s: %w", path, err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(file, picture); err != nil {
		file.Close()
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return file.Close()
}

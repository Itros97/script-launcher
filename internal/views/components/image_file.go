package components

import (
	"os"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// ImageFiles shows pictures from disk, decoding each once and again only when the file changes.
type ImageFiles struct {
	cache map[string]cachedImage
}

type cachedImage struct {
	picture  *draw2d.Image
	modified time.Time
}

// Draw shows the picture at path as big as fits in width × height, keeping its proportions; it
// returns false, drawing nothing, when the file is missing or is not a picture.
func (images *ImageFiles) Draw(context *ui.Context, path string, width, height float32) bool {
	picture := images.load(path)
	if picture == nil || width <= 0 || height <= 0 {
		return false
	}
	pixelsWide, pixelsHigh := picture.Size()
	scale := min(width/float32(pixelsWide), height/float32(pixelsHigh))
	context.Image(picture, ui.Layout{Width: ui.Fixed(float32(pixelsWide) * scale), Height: ui.Fixed(float32(pixelsHigh) * scale)}, 6)
	return true
}

// load returns the decoded picture, reading the file again when it changed since the last time.
func (images *ImageFiles) load(path string) *draw2d.Image {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if cached, ok := images.cache[path]; ok && cached.modified.Equal(info.ModTime()) {
		return cached.picture
	}
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	picture, err := draw2d.DecodeImage(file)
	if err != nil {
		return nil
	}
	if images.cache == nil {
		images.cache = map[string]cachedImage{}
	}
	images.cache[path] = cachedImage{picture: picture, modified: info.ModTime()}
	return picture
}

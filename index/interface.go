package index

import (
	"image"

	"github.com/timwu/mosaicer/analysis"
)

// Builder is an interface for building an index
type Builder interface {
	// Write the given image data into the index
	Index(name string, data *analysis.ImageData) error

	// Finish creating the index
	Close() error
}

// Index is an interface for wrapping up an image index for finding matching images
type Index interface {
	// Find the best matching image for the given source image
	Search(img *image.NRGBA, aspectRatio image.Point) (string, error)
}

package index

import "image"

// Index is an interface for wrapping up an image index for finding matching images
type Index interface {
	// Find the best matching image for the given source image
	Search(img *image.NRGBA, aspectRatio image.Point) (string, error)
}

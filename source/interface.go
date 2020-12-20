package source

import (
	"image"

	// Import image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var (
	imageFileTypes = map[string]bool{
		"gif":  true,
		"jpg":  true,
		"jpeg": true,
		"png":  true,
	}
)

// ImageSource is an interface for describing a source of images
type ImageSource interface {
	GetImageNames() ([]string, error)
	GetImage(name string) (image.Image, error)
	Close()
}

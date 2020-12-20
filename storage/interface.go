package storage

import "github.com/timwu/mosaicer/analysis"

// Storage stores metadata for images
type Storage interface {
	// Get the list of known keys
	Keys() ([]string, error)

	// Load image data with the given key
	Load(key string) (*analysis.ImageData, error)

	// Store an image with the given key for retrieval later
	Store(key string, data *analysis.ImageData) error
}

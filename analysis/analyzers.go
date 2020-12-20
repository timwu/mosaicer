package analysis

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/timwu/mosaicer/util"
)

// Simple performs a simple analysis of the given image into a 1x1 and aspect-ratio sized image samples
func Simple(img image.Image) (*ImageData, error) {
	data := &ImageData{
		AspectRatio: util.AspectRatio(img),
		Samples:     make([]*image.NRGBA, 0),
	}

	data.Samples = append(data.Samples, imaging.Resize(img, 1, 1, imaging.NearestNeighbor))
	data.Samples = append(data.Samples, imaging.Resize(img, data.AspectRatio.X, data.AspectRatio.Y, imaging.NearestNeighbor))
	return data, nil
}

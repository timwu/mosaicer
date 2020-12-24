package analysis

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/timwu/mosaicer/util"
)

const (
	samples = 2
)

// Simple performs a simple analysis of the given image into a 1x1 and aspect-ratio sized image samples
func Simple(img image.Image) (*ImageData, error) {
	data := &ImageData{
		AspectRatio: util.AspectRatio(img),
		Samples:     make([]*image.NRGBA, samples),
	}

	for i := 0; i < samples; i++ {
		size := data.AspectRatio.Mul(i)
		if i == 0 {
			size = image.Point{X: 1, Y: 1}
		}
		data.Samples[i] = imaging.Resize(img, size.X, size.Y, imaging.NearestNeighbor)
	}

	return data, nil
}

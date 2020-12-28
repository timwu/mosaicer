package analysis

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/timwu/mosaicer/util"
)

// Simple performs a simple analysis of the given image into a 1x1 and aspect-ratio sized image samples
func Simple(img image.Image, samples int) (*ImageData, error) {
	data := &ImageData{
		AspectRatio: util.AspectRatio(img),
		Samples:     make([]*image.NRGBA, 0),
		LabSamples:  make(map[image.Point][]float64),
	}

	// Skip images with weird aspect ratios (not 4:3)
	if data.AspectRatio.X*data.AspectRatio.Y > 150 {
		return data, nil
	}

	for i := 0; i < samples; i++ {
		size := data.AspectRatio.Mul(i)
		if i == 0 {
			size = image.Point{X: 1, Y: 1}
		}
		resized := imaging.Resize(img, size.X, size.Y, imaging.NearestNeighbor)
		data.Samples = append(data.Samples, resized)
		data.LabSamples[size] = RGBAToLab(resized.Pix)
	}

	return data, nil
}

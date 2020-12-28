package analysis

import "image"

type ImageData struct {
	AspectRatio image.Point
	Samples     []*image.NRGBA
	LabSamples  map[image.Point][]float64
}

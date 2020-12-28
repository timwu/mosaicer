package analysis

import "github.com/lucasb-eyer/go-colorful"

func bytesToColor(bytes []uint8) colorful.Color {
	return colorful.Color{R: float64(bytes[0]) / 255.0, G: float64(bytes[1]) / 255.0, B: float64(bytes[2]) / 255.0}
}

func RGBAToLab(bytes []uint8) []float64 {
	// shift from the rgba sample to l*a*b floats
	labColors := make([]float64, (len(bytes)/4)*3)
	for i := 0; i < len(bytes)/4; i++ {
		l, a, b := bytesToColor(bytes[i*4 : (i+1)*4]).Lab()
		labColors[i*3] = l
		labColors[i*3+1] = a
		labColors[i*3+2] = b
	}
	return labColors
}

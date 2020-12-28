package index

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type sample []uint8

func sq(f float64) float64 {
	return f * f
}

func basicDistance(left, right []uint8) float64 {
	redDiff := float64(left[0]) - float64(right[0])
	greenDiff := float64(left[1]) - float64(right[1])
	blueDiff := float64(left[2]) - float64(right[2])
	return math.Sqrt(redDiff*redDiff + greenDiff*greenDiff + blueDiff*blueDiff)
}

func redmeanDistance(left, right []uint8) float64 {
	redMean := (float64(left[0]) + float64(right[0])) / 2.0
	redDiff := float64(left[0]) - float64(right[0])
	greenDiff := float64(left[1]) - float64(right[1])
	blueDiff := float64(left[2]) - float64(right[2])

	return math.Sqrt((2.0+redMean/256.0)*redDiff*redDiff + 4*greenDiff*greenDiff + (2+(255.0-redMean)/256.0)*blueDiff*blueDiff)
}

func bytesToColor(bytes []uint8) colorful.Color {
	return colorful.Color{R: float64(bytes[0]) / 255.0, G: float64(bytes[1]) / 255.0, B: float64(bytes[2]) / 255.0}
}

func distance(left, right sample) float64 {
	var sumSquares float64
	for i := 0; i < len(left); i += 4 {
		leftColor := bytesToColor(left[i : i+4])
		rightColor := bytesToColor(right[i : i+4])
		d := leftColor.DistanceLab(rightColor)
		sumSquares += d
	}
	return sumSquares
}

func floatDistance(left, right []float64) float64 {
	var totalDistance float64
	for i := 0; i < len(left); i += 3 {
		totalDistance += math.Sqrt(sq(left[i]-right[i]) + sq(left[i+1]-right[i+1]) + sq(left[i+2]-right[i+2]))
	}
	return totalDistance / float64(len(left)/3)
}

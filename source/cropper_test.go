package source

import (
	"image"
	"testing"
)

func TestCrop(t *testing.T) {
	cropped := cropToAspectRatio(image.Point{
		X: 3072, Y: 4080,
	}, image.Point{X: 4, Y: 3})
	expected := image.Point{
		X: 3060, Y: 4080,
	}
	if cropped != expected {
		t.Fatalf("Got wrong cropped size %v, expected %v", cropped, expected)
	}
}

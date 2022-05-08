package util

import (
	"image"
	"testing"
)

func TestMinTiles(t *testing.T) {
	actual := MinTiles(image.Point{X: 1, Y: 1}, image.Point{X: 4, Y: 3})
	expected := image.Point{
		X: 3, Y: 4,
	}
	if actual != expected {
		t.Fatalf("Wrong number of tiles, got %v, expected %v", actual, expected)
	}

	actual = MinTiles(image.Point{16, 9}, image.Point{4, 3})
	expected = image.Point{4, 3}
	if actual != expected {
		t.Fatalf("Wrong number of tiles, got %v, expected %v", actual, expected)
	}
}

func TestConvertTiles(t *testing.T) {
	actual := ConvertTiles(image.Point{4, 3}, image.Point{4, 3}, 100)
	expected := image.Point{100, 100}
	if actual != expected {
		t.Fatalf("Wrong number of tiles actual=%v, expected=%v", actual, expected)
	}

	actual = ConvertTiles(image.Point{16, 9}, image.Point{4, 3}, 100)
	expected = image.Point{100, 75}
	if actual != expected {
		t.Fatalf("Wrong number of tiles actual=%v, expected=%v", actual, expected)
	}

	actual = ConvertTiles(image.Point{16, 9}, image.Point{4, 3}, 101)
	expected = image.Point{104, 78}
	if actual != expected {
		t.Fatalf("Wrong number of tiles actual=%v, expected=%v", actual, expected)
	}
}

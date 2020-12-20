package util

import "image"

// greatest common divisor (GCD) via Euclidean algorithm
func gcd(a, b int) int {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

// AspectRatio calculates the minimal aspect ratio of the given image
func AspectRatio(img image.Image) image.Point {
	size := img.Bounds().Size()
	divisor := gcd(size.X, size.Y)
	return image.Point{
		X: size.X / divisor,
		Y: size.Y / divisor,
	}
}

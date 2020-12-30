// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"image"
)

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

// Paste the src image into the dst image at the given location on the dst image. returns
// an error if the src image would not fit
func Paste(dst, src *image.NRGBA, loc image.Point) error {
	srcSize := src.Rect.Size()
	dstSize := dst.Rect.Size()

	if loc.X+srcSize.X > dstSize.X {
		return fmt.Errorf("too big in X dimension")
	}
	if loc.Y+srcSize.Y > dstSize.Y {
		return fmt.Errorf("too big in Y direction")
	}

	for row := 0; row < srcSize.Y; row++ {
		copy(dst.Pix[(loc.Y+row)*dst.Stride+(loc.X*4):(loc.Y+row)*dst.Stride+(loc.X*4)+(srcSize.X*4)],
			src.Pix[row*src.Stride:row*src.Stride+(srcSize.X*4)])
	}
	return nil
}

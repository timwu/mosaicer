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
	"math"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var (
	saneRatios = map[float64]image.Point{
		1.0:        image.Point{1, 1},
		4.0 / 3.0:  image.Point{4, 3},
		16.0 / 9.0: image.Point{16, 9},
	}
)

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

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
func Paste(dst, src *image.NRGBA, loc image.Point, blend float64) error {
	if blend <= 0.0 || blend > 1.0 {
		return fmt.Errorf("blend must be between (0.0, 1.0]")
	}

	srcSize := src.Rect.Size()
	dstSize := dst.Rect.Size()

	if loc.X+srcSize.X > dstSize.X {
		return fmt.Errorf("too big in X dimension")
	}
	if loc.Y+srcSize.Y > dstSize.Y {
		return fmt.Errorf("too big in Y direction")
	}

	if blend < 1.0 {
		background := imaging.Crop(dst, image.Rect(loc.X, loc.Y, loc.X+srcSize.X, loc.Y+srcSize.Y))
		src = imaging.Overlay(background, src, image.Point{0, 0}, blend)
	}

	for row := 0; row < srcSize.Y; row++ {
		copy(dst.Pix[(loc.Y+row)*dst.Stride+(loc.X*4):(loc.Y+row)*dst.Stride+(loc.X*4)+(srcSize.X*4)],
			src.Pix[row*src.Stride:row*src.Stride+(srcSize.X*4)])
	}

	return nil
}

// Calculate the minimum number of tiles in each direction to go from the tile aspect ratio to the image aspect ratio
func MinTiles(imageAspectRatio image.Point, tileAspectRatio image.Point) image.Point {
	target := image.Point{
		X: tileAspectRatio.Y * imageAspectRatio.X,
		Y: tileAspectRatio.X * imageAspectRatio.Y,
	}
	divisor := gcd(target.X, target.Y)
	return image.Point{
		X: target.X / divisor,
		Y: target.Y / divisor,
	}
}

func ConvertTiles(imageAspectRatio, tileAspectRatio image.Point, targetTileCount int) image.Point {
	minTiles := MinTiles(imageAspectRatio, tileAspectRatio)

	if minTiles.X > minTiles.Y {
		factor := int(math.Ceil(float64(targetTileCount) / float64(minTiles.X)))
		return image.Point{minTiles.X * factor, minTiles.Y * factor}
	} else {
		factor := int(math.Ceil(float64(targetTileCount) / float64(minTiles.Y)))
		return image.Point{minTiles.X * factor, minTiles.Y * factor}
	}
}

func ParseAspectRatioString(aspectRatio string) (image.Point, error) {
	split := strings.SplitN(aspectRatio, ":", 2)
	if len(split) != 2 {
		return image.Point{}, fmt.Errorf("invalid aspect ratio string %s", aspectRatio)
	}
	x, err := strconv.Atoi(split[0])
	if err != nil {
		return image.Point{}, err
	}
	y, err := strconv.Atoi(split[1])
	if err != nil {
		return image.Point{}, err
	}
	return image.Point{x, y}, nil
}

func NearestSaneAspectRatio(aspectRatio image.Point) image.Point {
	ratio := float64(max(aspectRatio.X, aspectRatio.Y)) / float64(min(aspectRatio.Y, aspectRatio.X))
	returnRatio := aspectRatio
	nearestDistance := math.MaxFloat64
	for r, p := range saneRatios {
		if distance := math.Abs(r - ratio); distance < nearestDistance {
			nearestDistance = distance
			returnRatio = p
		}
	}
	return returnRatio
}

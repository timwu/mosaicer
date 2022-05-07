package source

import (
	"image"

	"github.com/disintegration/imaging"
)

type cropSource struct {
	src               ImageSource
	targetAspectRatio image.Point
}

func NewCropSource(src ImageSource, targetAspectRatio image.Point) ImageSource {
	return &cropSource{src, targetAspectRatio}
}

func CropImageToAspectRatio(img image.Image, targetAspectRatio image.Point) image.Image {
	targetSize := cropToAspectRatio(img.Bounds().Size(), targetAspectRatio)
	return imaging.CropCenter(img, targetSize.X, targetSize.Y)
}

func cropToAspectRatio(size, targetAspectRatio image.Point) image.Point {
	if (size.X > size.Y) != (targetAspectRatio.X > targetAspectRatio.Y) {
		// Orientation does not match, rotate the target aspect ratio to match first
		targetAspectRatio = image.Point{X: targetAspectRatio.Y, Y: targetAspectRatio.X}
	}

	targetWidth := size.X - (size.X % targetAspectRatio.X)
	targetHeight := (targetWidth / targetAspectRatio.X) * targetAspectRatio.Y
	if targetHeight > size.Y {
		targetHeight = size.Y - (size.Y % targetAspectRatio.Y)
		targetWidth = (targetHeight / targetAspectRatio.Y) * targetAspectRatio.X
	}
	return image.Point{X: targetWidth, Y: targetHeight}
}

func (c *cropSource) GetImageNames() ([]string, error) {
	return c.src.GetImageNames()
}

func (c *cropSource) Close() {
	c.src.Close()
}

func (c *cropSource) GetImage(name string) (image.Image, error) {
	baseImg, err := c.src.GetImage(name)
	if err != nil {
		return nil, err
	}
	return CropImageToAspectRatio(baseImg, c.targetAspectRatio), nil
}

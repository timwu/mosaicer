package index

import (
	"fmt"
	"image"
	"math"

	"github.com/timwu/mosaicer/storage"
	"github.com/timwu/mosaicer/util"
)

type sample []uint8

type inMemoryIndex struct {
	keyToID  map[string]int
	idToKey  []string
	multiple int
	// all the sample data, maps from aspect ratio -> samples by id
	samples map[image.Point]map[int]sample
}

func toSample(img *image.NRGBA) sample {
	nBytes := img.Rect.Size().X * img.Rect.Size().Y * 4
	if len(img.Pix) == nBytes {
		return img.Pix
	}
	panic("Wrong byte size!")
}

func colorDistance(left, right []uint8) float64 {
	redMean := (float64(left[0]) + float64(right[0])) / 2.0
	redDiff := float64(left[0]) - float64(right[0])
	greenDiff := float64(left[1]) - float64(right[1])
	blueDiff := float64(left[2]) - float64(right[2])

	return math.Sqrt((2.0+redMean/256.0)*redDiff*redDiff + 4*greenDiff*greenDiff + (2+(255.0-redMean)/256.0)*blueDiff*blueDiff)
}

func distance(left, right sample) float64 {
	var sumSquares float64
	for i := 0; i < len(left); i += 4 {
		sumSquares += colorDistance(left[i:i+4], right[i:i+4])
	}
	return sumSquares
}

func (i *inMemoryIndex) Search(img *image.NRGBA, aspectRatio image.Point) (string, error) {
	// defer util.LogTime("search")()
	size := img.Rect.Size()
	if i.multiple == 0 {
		if size.X != 1 || size.Y != 1 {
			return "", fmt.Errorf("incorrectly sized image. Expecting 1x1. Got %v", size)
		}
	} else if (i.multiple*aspectRatio.X != size.X) || (i.multiple*aspectRatio.Y != size.Y) {
		return "", fmt.Errorf("incorrectly sized image")
	}

	testSample := toSample(img)
	shortestDistance := math.MaxFloat64
	closestID := -1
	for id := range i.samples[aspectRatio] {
		d := distance(i.samples[aspectRatio][id], testSample)
		if d < shortestDistance {
			shortestDistance = d
			closestID = id
		}
	}

	if closestID == -1 {
		return "", fmt.Errorf("could not find any samples")
	}

	return i.idToKey[closestID], nil
}

// BuildInMemoryIndex builds an in memory index of the image samples at the given multiple of the aspect ratio. 0 is special in that it is a 1x1.
func BuildInMemoryIndex(storage storage.Storage, multiple int) (Index, error) {
	defer util.LogTime("build index")()
	keys, err := storage.Keys()
	if err != nil {
		return nil, err
	}
	index := &inMemoryIndex{
		keyToID:  make(map[string]int),
		idToKey:  keys,
		multiple: multiple,
		samples:  make(map[image.Point]map[int]sample),
	}
	for id, key := range keys {
		index.keyToID[key] = id
		data, err := storage.Load(key)
		if err != nil {
			return nil, err
		}
		if index.samples[data.AspectRatio] == nil {
			index.samples[data.AspectRatio] = make(map[int]sample)
		}
		// The requested multiple is not available, error out
		if len(data.Samples) < (multiple + 1) {
			return nil, fmt.Errorf("requested multiple %d not available for %s", multiple, key)
		}
		index.samples[data.AspectRatio][id] = toSample(data.Samples[multiple])
	}
	return index, nil
}

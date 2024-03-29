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

package index

import (
	"fmt"
	"image"
	"math/rand"
	"sort"

	"github.com/disintegration/imaging"
	"github.com/timwu/mosaicer/analysis"
	bolt "go.etcd.io/bbolt"
)

// the data hierarchy in the bolt db is:
// v1
// - names
//   - int key -> string name
// - data
//   - dimensions
//     - int key -> rgba bytes
// - lab_data
//   - dimensions
//     - int key -> lab bytes
var (
	indexSuffix = ".index.bolt"

	rootKey    = []byte("v1")
	namesKey   = []byte("names")
	dataKey    = []byte("data")
	labDataKey = []byte("lab_data")
)

func boltDB(source string) (*bolt.DB, error) {
	return bolt.Open(source+indexSuffix, 0666, nil)
}

type boltIndexBuilder struct {
	db *bolt.DB
}

func addName(name string, rootBucket *bolt.Bucket) (int, error) {
	namesBucket, err := rootBucket.CreateBucketIfNotExists(namesKey)
	if err != nil {
		return 0, err
	}
	id, err := namesBucket.NextSequence()
	if err != nil {
		return 0, err
	}
	// There can be duplicates, just don't repeat the same dataset?
	return int(id), namesBucket.Put(intToBytes(int(id)), []byte(name))
}

func (b *boltIndexBuilder) Index(name string, data *analysis.ImageData) error {
	// Don't bother storing images with no samples
	if len(data.Samples) == 0 {
		return nil
	}
	return b.db.Batch(func(tx *bolt.Tx) error {
		rootBucket, err := tx.CreateBucketIfNotExists(rootKey)
		if err != nil {
			return err
		}
		id, err := addName(name, rootBucket)
		if err != nil {
			return err
		}
		dataBucket, err := rootBucket.CreateBucketIfNotExists(dataKey)
		if err != nil {
			return err
		}
		for _, sample := range data.Samples {
			dimensionBucket, err := dataBucket.CreateBucketIfNotExists(pointToBytes(sample.Rect.Size()))
			if err != nil {
				return err
			}
			if err := dimensionBucket.Put(intToBytes(id), sample.Pix); err != nil {
				return err
			}
		}
		labDataBucket, err := rootBucket.CreateBucketIfNotExists(labDataKey)
		if err != nil {
			return err
		}
		for size, sample := range data.LabSamples {
			dimensionBucket, err := labDataBucket.CreateBucketIfNotExists(pointToBytes(size))
			if err != nil {
				return err
			}
			if err := dimensionBucket.Put(intToBytes(id), floatsToBytes(sample)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *boltIndexBuilder) Close() error {
	return b.db.Close()
}

// NewBoltIndexBuilder Creates a Bolt index builder
func NewBoltIndexBuilder(source string) (Builder, error) {
	db, err := boltDB(source)
	if err != nil {
		return nil, err
	}
	builder := &boltIndexBuilder{
		db: db,
	}
	return builder, nil
}

type boltIndex struct {
	db        *bolt.DB
	multiple  int
	fuzziness int
}

func getName(id int, rootBucket *bolt.Bucket) (string, error) {
	namesBucket := rootBucket.Bucket(namesKey)
	if namesBucket == nil {
		return "", fmt.Errorf("names bucket not found")
	}
	name := namesBucket.Get(intToBytes(id))
	if name == nil {
		return "", fmt.Errorf("name not found for id %d", id)
	}
	return string(name), nil
}

func getDistances(dataBucket *bolt.Bucket, size image.Point, bytes []byte, idDistances map[int]float64) error {
	query := analysis.RGBAToLab(bytes)
	dimensionBucket := dataBucket.Bucket(pointToBytes(size))
	if dimensionBucket == nil {
		return fmt.Errorf("dimension bucket not found: %v", size)
	}
	return dimensionBucket.ForEach(func(k, v []byte) error {
		id := bytesToInt(k)
		idDistances[id] = floatDistance(query, bytesToFloats(v))
		return nil
	})
}

func (b *boltIndex) Search(img *image.NRGBA, aspectRatio image.Point) (string, error) {
	selected := ""
	size := aspectRatio.Mul(b.multiple)
	if b.multiple == 0 {
		size = image.Point{X: 1, Y: 1}
	}
	resized := imaging.Resize(img, size.X, size.Y, imaging.NearestNeighbor)
	if err := b.db.View(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket(rootKey)
		if rootBucket == nil {
			return fmt.Errorf("root bucket not found")
		}
		dataBucket := rootBucket.Bucket(labDataKey)
		if dataBucket == nil {
			return fmt.Errorf("data bucket not found")
		}

		idDistances := make(map[int]float64)
		if err := getDistances(dataBucket, size, resized.Pix, idDistances); err != nil {
			return err
		}
		// If the image is not square, also consider the rotated version
		if size.X != size.Y {
			rotated := imaging.Rotate90(resized)
			if err := getDistances(dataBucket, rotated.Rect.Size(), rotated.Pix, idDistances); err != nil {
				return err
			}
		}

		ids := make([]int, len(idDistances))
		i := 0
		for id := range idDistances {
			ids[i] = id
			i++
		}
		sort.Slice(ids, func(i, j int) bool {
			return idDistances[ids[i]] < idDistances[ids[j]]
		})

		var err error
		selected, err = getName(ids[rand.Intn(b.fuzziness)], rootBucket)
		return err
	}); err != nil {
		return "", err
	}
	if selected == "" {
		return selected, fmt.Errorf("no matching image found")
	}
	return selected, nil
}

// NewBoltIndex creates a bolt index for searching
func NewBoltIndex(source string, multiple, fuzziness int) (Index, error) {
	db, err := boltDB(source)
	if err != nil {
		return nil, err
	}
	index := &boltIndex{
		db:        db,
		multiple:  multiple,
		fuzziness: fuzziness,
	}
	return index, nil
}

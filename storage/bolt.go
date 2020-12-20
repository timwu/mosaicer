package storage

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/timwu/mosaicer/analysis"
	bolt "go.etcd.io/bbolt"
)

const (
	dbSuffix    = ".storagedb"
	resizeDepth = 1
)

var (
	rootBucketKey  = []byte("v1")
	aspectRatioKey = "aspect_ratio"
	samplesKey     = "sample"

	// image.ImageNRBA fields
	pixKey    = "pix"
	strideKey = "stride"
	minKey    = "min"
	maxKey    = "max"
)

type boltStorage struct {
	db *bolt.DB
}

func storePoint(key string, point image.Point, bucket *bolt.Bucket) error {
	pointBucket, err := bucket.CreateBucketIfNotExists([]byte(key))
	if err != nil {
		return err
	}
	if err := pointBucket.Put([]byte("x"), itob(point.X)); err != nil {
		return err
	}
	if err := pointBucket.Put([]byte("y"), itob(point.Y)); err != nil {
		return err
	}
	return nil
}

func loadPoint(key string, bucket *bolt.Bucket) (image.Point, error) {
	pointBucket := bucket.Bucket([]byte(key))
	if pointBucket == nil {
		return image.Point{}, fmt.Errorf("point not found")
	}
	xBytes := pointBucket.Get([]byte("x"))
	if xBytes == nil {
		return image.Point{}, fmt.Errorf("can't find x")
	}
	yBytes := pointBucket.Get([]byte("y"))
	if yBytes == nil {
		return image.Point{}, fmt.Errorf("can't find y")
	}
	return image.Point{
		X: btoi(xBytes),
		Y: btoi(yBytes),
	}, nil
}

func storageImage(key int, img *image.NRGBA, bucket *bolt.Bucket) error {
	imageBucket, err := bucket.CreateBucketIfNotExists(itob(key))
	if err != nil {
		return err
	}
	if err := imageBucket.Put([]byte(pixKey), img.Pix); err != nil {
		return err
	}
	if err := imageBucket.Put([]byte(strideKey), itob(img.Stride)); err != nil {
		return err
	}
	if err := storePoint(minKey, img.Rect.Min, imageBucket); err != nil {
		return err
	}
	if err := storePoint(maxKey, img.Rect.Max, imageBucket); err != nil {
		return err
	}
	return nil
}

func copyBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func loadImage(key int, bucket *bolt.Bucket) (*image.NRGBA, error) {
	imageBucket := bucket.Bucket(itob(key))
	if imageBucket == nil {
		return nil, fmt.Errorf("image not found")
	}
	min, err := loadPoint(minKey, imageBucket)
	if err != nil {
		return nil, err
	}
	max, err := loadPoint(maxKey, imageBucket)
	if err != nil {
		return nil, err
	}
	strideBytes := imageBucket.Get([]byte(strideKey))
	if strideBytes == nil {
		return nil, fmt.Errorf("missing stride")
	}
	return &image.NRGBA{
		Pix:    copyBytes(imageBucket.Get([]byte(pixKey))),
		Stride: btoi(strideBytes),
		Rect: image.Rectangle{
			Min: min,
			Max: max,
		},
	}, nil
}

func (b *boltStorage) Keys() ([]string, error) {
	var returnKeys []string
	if err := b.db.View(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket(rootBucketKey)
		if rootBucket == nil {
			return fmt.Errorf("root bucket not found")
		}
		keys := make([]string, 0)
		if err := rootBucket.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		}); err != nil {
			return err
		}
		returnKeys = keys
		return nil
	}); err != nil {
		return nil, err
	}
	return returnKeys, nil
}

func (b *boltStorage) Load(key string) (*analysis.ImageData, error) {
	data := &analysis.ImageData{
		Samples: make([]*image.NRGBA, 0),
	}
	if err := b.db.View(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket(rootBucketKey)
		if rootBucket == nil {
			return fmt.Errorf("root bucket not found")
		}
		imageBucket := rootBucket.Bucket([]byte(key))
		if imageBucket == nil {
			return fmt.Errorf("image not found %s", key)
		}
		var err error
		data.AspectRatio, err = loadPoint(aspectRatioKey, imageBucket)
		if err != nil {
			return err
		}
		samplesBucket := imageBucket.Bucket([]byte(samplesKey))
		samples := make([]*image.NRGBA, 0)
		if err := samplesBucket.ForEach(func(k, v []byte) error {
			img, err := loadImage(btoi(k), samplesBucket)
			if err != nil {
				return err
			}
			samples = append(samples, img)
			return nil
		}); err != nil {
			return err
		}
		data.Samples = samples
		return nil
	}); err != nil {
		return nil, err
	}
	return data, nil
}

func (b *boltStorage) Store(key string, data *analysis.ImageData) error {
	return b.db.Batch(func(tx *bolt.Tx) error {
		rootBucket, err := tx.CreateBucketIfNotExists(rootBucketKey)
		if err != nil {
			return err
		}
		imageBucket, err := rootBucket.CreateBucketIfNotExists([]byte(key))
		if err != nil {
			return err
		}
		if err := storePoint(aspectRatioKey, data.AspectRatio, imageBucket); err != nil {
			return err
		}
		samplesBucket, err := imageBucket.CreateBucketIfNotExists([]byte(samplesKey))
		if err != nil {
			return nil
		}
		for index, img := range data.Samples {
			if err := storageImage(index, img, samplesBucket); err != nil {
				return err
			}
		}
		return nil
	})
}

// NewBoltStorage get a bolt index for the given target
func NewBoltStorage(target string) (Storage, error) {
	db, err := bolt.Open(target+dbSuffix, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &boltStorage{
		db: db,
	}, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(v []byte) int {
	return int(binary.BigEndian.Uint64(v))
}

package source

import (
	"archive/zip"
	"fmt"
	"image"
	"path"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
)

const (
	cacheTTL = 30 * time.Second
)

type zipImageSource struct {
	reader *zip.ReadCloser
	images map[string]*zip.File
	cache  *ttlcache.Cache
}

func (z *zipImageSource) GetImageNames() ([]string, error) {
	names := make([]string, 0)
	for name := range z.images {
		names = append(names, name)
	}
	return names, nil
}

func (z *zipImageSource) GetImage(name string) (image.Image, error) {
	// defer util.LogTime(fmt.Sprintf("Load %s from zip.", name))()
	img, err := z.cache.Get(name)
	return img.(image.Image), err
}

func (z *zipImageSource) Close() {
	z.reader.Close()
}

// NewZipImageSource creates an ImageSource from the given zip file
func NewZipImageSource(zipFile string) (ImageSource, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	z := &zipImageSource{
		reader: r,
		images: make(map[string]*zip.File),
		cache:  ttlcache.NewCache(),
	}

	for _, f := range r.File {
		if imageFileTypes[strings.TrimPrefix(path.Ext(f.Name), ".")] {
			z.images[f.Name] = f
		}
	}

	z.cache.SetLoaderFunction(func(key string) (data interface{}, ttl time.Duration, err error) {
		if f := z.images[key]; f != nil {
			r, err := f.Open()
			if err != nil {
				return nil, cacheTTL, err
			}
			defer r.Close()
			img, _, err := image.Decode(r)
			return img, cacheTTL, err
		}
		return nil, cacheTTL, fmt.Errorf("image %s not found", key)
	})
	return z, nil
}

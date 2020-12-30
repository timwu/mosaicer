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

package source

import (
	"archive/zip"
	"fmt"
	"image"
	"path"
	"strings"
)

type zipImageSource struct {
	reader *zip.ReadCloser
	images map[string]*zip.File
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
	if f := z.images[name]; f != nil {
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		img, _, err := image.Decode(r)
		return img, err
	}
	return nil, fmt.Errorf("image not found %s", name)
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
	}

	for _, f := range r.File {
		if imageFileTypes[strings.TrimPrefix(path.Ext(f.Name), ".")] {
			z.images[f.Name] = f
		}
	}
	return z, nil
}

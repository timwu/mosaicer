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
	"image"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type folderImageSource struct {
	dir string
}

func (f folderImageSource) GetImageNames() ([]string, error) {
	fileInfos, err := ioutil.ReadDir(f.dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		if imageFileTypes[strings.TrimPrefix(path.Ext(fileInfo.Name()), ".")] {
			names = append(names, fileInfo.Name())
		}
	}
	return names, nil
}

func (f folderImageSource) GetImage(name string) (image.Image, error) {
	r, err := os.Open(path.Join(f.dir, name))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	img, _, err := image.Decode(r)
	return img, err
}

func (f folderImageSource) Close() {}

// NewFolderImageSource creates a folder-backed ImageSource
func NewFolderImageSource(dir string) (ImageSource, error) {
	return folderImageSource{
		dir: dir,
	}, nil
}

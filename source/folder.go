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
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type folderImageSource struct {
	dir string
}

func joinZipFileName(zipFileName, fileName string) string {
	return zipFileName + "||" + fileName
}

func splitZipFileName(joinedName string) (string, string, error) {
	splitName := strings.SplitN(joinedName, "||", 2)
	if len(splitName) != 2 {
		return "", "", fmt.Errorf("invalid zip file name to be split")
	}
	return splitName[0], splitName[1], nil
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
		extension := strings.TrimPrefix(path.Ext(fileInfo.Name()), ".")
		if imageFileTypes[extension] {
			names = append(names, fileInfo.Name())
		} else if extension == "zip" {
			zipSource, err := NewZipImageSource(path.Join(f.dir, fileInfo.Name()))
			if err != nil {
				return nil, err
			}
			defer zipSource.Close()
			zipImageNames, err := zipSource.GetImageNames()
			if err != nil {
				return nil, err
			}
			for i, n := range zipImageNames {
				zipImageNames[i] = joinZipFileName(fileInfo.Name(), n)
			}
			names = append(names, zipImageNames...)
		}
	}
	return names, nil
}

func (f folderImageSource) GetImage(name string) (image.Image, error) {
	if zipFileName, imageFileName, err := splitZipFileName(name); err == nil {
		zipImageSource, err := NewZipImageSource(path.Join(f.dir, zipFileName))
		if err != nil {
			return nil, err
		}
		defer zipImageSource.Close()
		return zipImageSource.GetImage(imageFileName)
	}
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

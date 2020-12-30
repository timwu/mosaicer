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

	// Import image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var (
	imageFileTypes = map[string]bool{
		"gif":  true,
		"jpg":  true,
		"jpeg": true,
		"png":  true,
	}
)

// ImageSource is an interface for describing a source of images
type ImageSource interface {
	GetImageNames() ([]string, error)
	GetImage(name string) (image.Image, error)
	Close()
}

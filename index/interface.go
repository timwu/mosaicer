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
	"image"

	"github.com/timwu/mosaicer/analysis"
)

// Builder is an interface for building an index
type Builder interface {
	// Write the given image data into the index
	Index(name string, data *analysis.ImageData) error

	// Finish creating the index
	Close() error
}

// Index is an interface for wrapping up an image index for finding matching images
type Index interface {
	// Find the best matching image for the given source image
	Search(img *image.NRGBA, aspectRatio image.Point) (string, error)
}

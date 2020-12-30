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

package storage

import "github.com/timwu/mosaicer/analysis"

// Storage stores metadata for images
type Storage interface {
	// Get the list of known keys
	Keys() ([]string, error)

	// Load image data with the given key
	Load(key string) (*analysis.ImageData, error)

	// Store an image with the given key for retrieval later
	Store(key string, data *analysis.ImageData) error
}

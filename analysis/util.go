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

package analysis

import "github.com/lucasb-eyer/go-colorful"

func bytesToColor(bytes []uint8) colorful.Color {
	return colorful.Color{R: float64(bytes[0]) / 255.0, G: float64(bytes[1]) / 255.0, B: float64(bytes[2]) / 255.0}
}

func RGBAToLab(bytes []uint8) []float64 {
	// shift from the rgba sample to l*a*b floats
	labColors := make([]float64, (len(bytes)/4)*3)
	for i := 0; i < len(bytes)/4; i++ {
		l, a, b := bytesToColor(bytes[i*4 : (i+1)*4]).Lab()
		labColors[i*3] = l
		labColors[i*3+1] = a
		labColors[i*3+2] = b
	}
	return labColors
}

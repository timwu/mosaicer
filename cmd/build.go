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

package cmd

import (
	"image"
	"image/color"
	"log"
	"os"
	"runtime/pprof"

	"github.com/cheggaaa/pb/v3"
	"github.com/disintegration/imaging"
	"github.com/spf13/cobra"
	"github.com/timwu/mosaicer/index"
	"github.com/timwu/mosaicer/source"
	"github.com/timwu/mosaicer/util"
)

var (
	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build photo mosaic output",
		RunE:  doBuild,
	}

	src                    = ""
	tiles                  = 0
	tileMultiple           = 20
	tileAspectRatio        = image.Point{4, 3}
	fuzziness              = 0
	referencePatchMultiple = 1
	cpuprofile             = ""
	tileSelectionThreads   = 10
	tilingThreads          = 16

	cropImageAspectRatio = ""
)

func init() {
	buildCmd.Flags().StringVar(&src, "source", "", "image source. must already have a built storage")
	buildCmd.Flags().IntVar(&tiles, "tiles", 100, "number of tiles in each dimension")
	buildCmd.Flags().IntVar(&fuzziness, "fuzziness", 5, "number of top images to consider for random selection")
	buildCmd.Flags().IntVar(&referencePatchMultiple, "referencePatchMultiple", 2, "Multiple of the aspect ratio for sizing a patch of the reference image")
	buildCmd.Flags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	buildCmd.Flags().StringVar(&cropImageAspectRatio, "cropImageAspectRatio", "auto", "Aspect ratio to crop the target image to before tiling.")
	rootCmd.AddCommand(buildCmd)
}

func selectImages(imgIndex index.Index, targetImg image.Image) (map[string][]image.Point, image.Point, error) {
	referencePatchSize := tileAspectRatio.Mul(referencePatchMultiple)

	imageAspectRatio := util.AspectRatio(targetImg)
	log.Printf("target img aspect ratio %v, base resolution of %v. Tile aspect ratio %v", imageAspectRatio, targetImg.Bounds().Size(), tileAspectRatio)

	tileCount := util.ConvertTiles(imageAspectRatio, tileAspectRatio, tiles)
	log.Printf("tile count %v", tileCount)
	referenceImg := imaging.Resize(targetImg, tileCount.X*referencePatchSize.X, 0, imaging.NearestNeighbor)
	log.Printf("reference img aspect ratio %v, size %v", util.AspectRatio(referenceImg), referenceImg.Rect.Size())

	progressBar := pb.StartNew(tileCount.X * tileCount.Y)
	tileNames := make(map[string][]image.Point)

	type tileSelection struct {
		selectedImage string
		point         image.Point
	}
	selectionsChan := make(chan tileSelection, 10)
	done := make(chan bool)
	go func() {
		for i := 0; i < tileCount.X*tileCount.Y; i++ {
			t := <-selectionsChan
			if tileNames[t.selectedImage] == nil {
				tileNames[t.selectedImage] = make([]image.Point, 0)
			}
			tileNames[t.selectedImage] = append(tileNames[t.selectedImage], t.point)

			progressBar.Increment()
		}
		progressBar.Finish()
		done <- true
	}()

	log.Printf("selecting images for tiles")
	limiter := util.NewLimiter(tileSelectionThreads)
	for i := 0; i < tileCount.Y; i++ {
		for j := 0; j < tileCount.X; j++ {
			i, j := i, j
			limiter.Go(func() {
				clip := imaging.Crop(referenceImg, image.Rectangle{
					Min: image.Point{X: j * referencePatchSize.X, Y: i * referencePatchSize.Y},
					Max: image.Point{X: (j + 1) * referencePatchSize.X, Y: (i + 1) * referencePatchSize.Y},
				})
				selected, err := imgIndex.Search(clip, tileAspectRatio)
				if err != nil {
					log.Fatalf("i=%d,j=%d err=%v", i, j, err)
				}
				selectionsChan <- tileSelection{
					selectedImage: selected,
					point:         image.Point{X: j, Y: i},
				}
			})
		}
	}
	limiter.Close()
	<-done
	return tileNames, tileCount, nil
}

func createOutputImage(imageSource source.ImageSource, tileNames map[string][]image.Point, tileCount image.Point) (*image.NRGBA, error) {
	log.Printf("Building output image")
	tileSize := tileAspectRatio.Mul(tileMultiple)
	dstImgSize := image.Point{X: tileSize.X * tileCount.X, Y: tileSize.Y * tileCount.Y}
	log.Printf("dst img size %v", dstImgSize)
	dstImg := imaging.New(dstImgSize.X, dstImgSize.Y, color.NRGBA{0, 0, 0, 0})
	progressBar := pb.StartNew(tileCount.X * tileCount.Y)
	rotatedTiles := 0
	limiter := util.NewLimiter(tilingThreads)
	for selectedName, points := range tileNames {
		selectedName, points := selectedName, points
		limiter.Go(func() {
			selectedImg, err := imageSource.GetImage(selectedName)
			if err != nil {
				log.Fatal(err)
			}

			// If the image is rotated relative to the target image's aspect ratio, rotate it first
			if ar := util.AspectRatio(selectedImg); ar.X == tileAspectRatio.Y && ar.Y == tileAspectRatio.X {
				selectedImg = imaging.Rotate270(selectedImg)
				rotatedTiles++
			}

			resizedTile := imaging.Resize(selectedImg, tileSize.X, 0, imaging.NearestNeighbor)

			for _, point := range points {
				if err := util.Paste(dstImg, resizedTile, image.Point{X: point.X * tileSize.X, Y: point.Y * tileSize.Y}); err != nil {
					log.Fatal(err)
				}
				progressBar.Increment()
			}
		})
	}
	limiter.Close()
	progressBar.Finish()
	log.Printf("Used %d rotated tiles", rotatedTiles)
	return dstImg, nil
}

func doBuild(cmd *cobra.Command, args []string) error {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	imageSource, err := source.NewImageSource(src)
	if err != nil {
		return err
	}
	imageSource = source.NewCropSource(imageSource, image.Point{X: 4, Y: 3})
	defer imageSource.Close()
	imgIndex, err := index.NewBoltIndex(src, referencePatchMultiple, fuzziness)
	if err != nil {
		return err
	}
	targetImg, err := imaging.Open(args[0])
	if err != nil {
		return err
	}

	if cropImageAspectRatio == "auto" {
		targetImg = source.CropImageToAspectRatio(targetImg, util.NearestSaneAspectRatio(util.AspectRatio(targetImg)))
	} else if cropImageAspectRatio != "none" {
		croppedAspectRatio, err := util.ParseAspectRatioString(cropImageAspectRatio)
		if err != nil {
			return err
		}
		targetImg = source.CropImageToAspectRatio(targetImg, croppedAspectRatio)
	}

	tileNames, tileCount, err := selectImages(imgIndex, targetImg)
	if err != nil {
		return err
	}

	log.Printf("Used %d unique images.", len(tileNames))
	dstImg, err := createOutputImage(imageSource, tileNames, tileCount)
	if err != nil {
		return err
	}
	imaging.Save(dstImg, args[0]+".mosaic.jpg")
	return nil
}

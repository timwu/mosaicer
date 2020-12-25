package cmd

import (
	"image"
	"image/color"
	"log"

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
	fuzziness              = 0
	referencePatchMultiple = 1
)

func init() {
	buildCmd.Flags().StringVar(&src, "source", "", "image source. must already have a built storage")
	buildCmd.Flags().IntVar(&tiles, "tiles", 100, "number of tiles in each dimension")
	buildCmd.Flags().IntVar(&fuzziness, "fuzziness", 5, "number of top images to consider for random selection")
	buildCmd.Flags().IntVar(&referencePatchMultiple, "referencePatchMultiple", 2, "Multiple of the aspect ratio for sizing a patch of the reference image")
	rootCmd.AddCommand(buildCmd)
}

func selectImages(imgIndex index.Index, targetImg image.Image, aspectRatio image.Point) (map[string][]image.Point, error) {
	referencePatchSize := aspectRatio.Mul(referencePatchMultiple)
	log.Printf("target img aspect ratio %v, base resolution of %v", aspectRatio, targetImg.Bounds().Size())
	referenceImg := imaging.Resize(targetImg, tiles*referencePatchSize.X, 0, imaging.NearestNeighbor)
	log.Printf("reference img aspect ratio %v, size %v", util.AspectRatio(referenceImg), referenceImg.Rect.Size())

	progressBar := pb.StartNew(tiles * tiles)
	tileNames := make(map[string][]image.Point)

	type tileSelection struct {
		selectedImage string
		point         image.Point
	}
	selectionsChan := make(chan tileSelection, 10)
	done := make(chan bool)
	go func() {
		for i := 0; i < tiles*tiles; i++ {
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
	for i := 0; i < tiles; i++ {
		for j := 0; j < tiles; j++ {
			i, j := i, j
			go func() {
				clip := imaging.Crop(referenceImg, image.Rectangle{
					Min: image.Point{X: j * referencePatchSize.X, Y: i * referencePatchSize.Y},
					Max: image.Point{X: (j + 1) * referencePatchSize.X, Y: (i + 1) * referencePatchSize.Y},
				})
				selected, err := imgIndex.Search(clip, aspectRatio)
				if err != nil {
					log.Fatal(err)
				}
				selectionsChan <- tileSelection{
					selectedImage: selected,
					point:         image.Point{X: j, Y: i},
				}
			}()
		}
	}
	<-done
	return tileNames, nil
}

func createOutputImage(imageSource source.ImageSource, tileNames map[string][]image.Point, aspectRatio image.Point) (*image.NRGBA, error) {
	log.Printf("Building output image")
	tileSize := aspectRatio.Mul(tileMultiple)
	dstImgSize := tileSize.Mul(tiles)
	log.Printf("dst img size %v", dstImgSize)
	dstImg := imaging.New(dstImgSize.X, dstImgSize.Y, color.NRGBA{0, 0, 0, 0})
	progressBar := pb.StartNew(tiles * tiles)
	rotatedTiles := 0
	for selectedName, points := range tileNames {
		selectedImg, err := imageSource.GetImage(selectedName)
		if err != nil {
			return nil, err
		}

		// If the image is rotated relative to the target image's aspect ratio, rotate it first
		if ar := util.AspectRatio(selectedImg); ar.X == aspectRatio.Y && ar.Y == aspectRatio.X {
			selectedImg = imaging.Rotate270(selectedImg)
			rotatedTiles++
		}

		resizedTile := imaging.Resize(selectedImg, tileSize.X, 0, imaging.NearestNeighbor)

		for _, point := range points {
			if err := util.Paste(dstImg, resizedTile, image.Point{X: point.X * tileSize.X, Y: point.Y * tileSize.Y}); err != nil {
				return nil, err
			}
			progressBar.Increment()
		}
	}
	progressBar.Finish()
	log.Printf("Used %d rotated tiles", rotatedTiles)
	return dstImg, nil
}

func doBuild(cmd *cobra.Command, args []string) error {
	imageSource, err := source.NewImageSource(src)
	if err != nil {
		return err
	}
	defer imageSource.Close()
	imgIndex, err := index.NewBoltIndex(src, referencePatchMultiple, fuzziness)
	if err != nil {
		return err
	}
	targetImg, err := imaging.Open(args[0])
	if err != nil {
		return err
	}

	aspectRatio := util.AspectRatio(targetImg)
	tileNames, err := selectImages(imgIndex, targetImg, aspectRatio)
	if err != nil {
		return err
	}

	log.Printf("Used %d unique images.", len(tileNames))
	dstImg, err := createOutputImage(imageSource, tileNames, aspectRatio)
	if err != nil {
		return err
	}
	imaging.Save(dstImg, args[0]+".mosaic.jpg")
	return nil
}

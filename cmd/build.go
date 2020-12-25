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

	src          = ""
	tiles        = 0
	tileMultiple = 20
	fuzziness    = 0
)

func init() {
	buildCmd.Flags().StringVar(&src, "source", "", "image source. must already have a built storage")
	buildCmd.Flags().IntVar(&tiles, "tiles", 100, "number of tiles in each dimension")
	buildCmd.Flags().IntVar(&fuzziness, "fuzziness", 5, "number of top images to consider for random selection")
	rootCmd.AddCommand(buildCmd)
}

func doBuild(cmd *cobra.Command, args []string) error {
	imageSource, err := source.NewImageSource(src)
	if err != nil {
		return err
	}
	defer imageSource.Close()
	imgIndex, err := index.NewBoltIndex(src, 1, fuzziness)
	if err != nil {
		return err
	}
	targetImg, err := imaging.Open(args[0])
	if err != nil {
		return err
	}

	aspectRatio := util.AspectRatio(targetImg)
	log.Printf("target img aspect ratio %v, base resolution of %v", aspectRatio, targetImg.Bounds().Size())
	referenceImg := imaging.Resize(targetImg, tiles*aspectRatio.X, 0, imaging.NearestNeighbor)
	log.Printf("reference img aspect ratio %v, size %v", util.AspectRatio(referenceImg), referenceImg.Rect.Size())

	progressBar := pb.StartNew(tiles * tiles)
	tileNames := make(map[string][]image.Point)
	log.Printf("selecting images for tiles")
	for i := 0; i < tiles; i++ {
		for j := 0; j < tiles; j++ {
			clip := imaging.Crop(referenceImg, image.Rectangle{
				Min: image.Point{X: j * aspectRatio.X, Y: i * aspectRatio.Y},
				Max: image.Point{X: (j + 1) * aspectRatio.X, Y: (i + 1) * aspectRatio.Y},
			})
			selected, err := imgIndex.Search(imaging.Resize(clip, aspectRatio.X, aspectRatio.Y, imaging.NearestNeighbor), aspectRatio)
			if err != nil {
				return err
			}
			if tileNames[selected] == nil {
				tileNames[selected] = make([]image.Point, 0)
			}
			tileNames[selected] = append(tileNames[selected], image.Point{X: j, Y: i})

			progressBar.Increment()
		}
	}
	progressBar.Finish()

	log.Printf("Used %d unique images.", len(tileNames))

	log.Printf("Building output image")
	tileSize := aspectRatio.Mul(tileMultiple)
	dstImgSize := tileSize.Mul(tiles)
	log.Printf("dst img size %v", dstImgSize)
	dstImg := imaging.New(dstImgSize.X, dstImgSize.Y, color.NRGBA{0, 0, 0, 0})
	progressBar = pb.StartNew(tiles * tiles)

	for selectedName, points := range tileNames {
		selectedImg, err := imageSource.GetImage(selectedName)
		if err != nil {
			return err
		}

		// If the image is rotated relative to the target image's aspect ratio, rotate it first
		if ar := util.AspectRatio(selectedImg); ar.X == aspectRatio.Y && ar.Y == aspectRatio.X {
			selectedImg = imaging.Rotate90(selectedImg)
		}

		resizedTile := imaging.Resize(selectedImg, tileSize.X, 0, imaging.NearestNeighbor)

		for _, point := range points {
			if err := util.Paste(dstImg, resizedTile, image.Point{X: point.X * tileSize.X, Y: point.Y * tileSize.Y}); err != nil {
				return err
			}
			progressBar.Increment()
		}
	}
	progressBar.Finish()

	imaging.Save(dstImg, args[0]+".mosaic.jpg")
	return nil
}

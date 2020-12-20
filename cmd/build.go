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
	"github.com/timwu/mosaicer/storage"
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
)

func init() {
	buildCmd.Flags().StringVar(&src, "source", "", "image source. must already have a built storage")
	buildCmd.Flags().IntVar(&tiles, "tiles", 100, "number of tiles in each dimension")
	rootCmd.AddCommand(buildCmd)
}

func doBuild(cmd *cobra.Command, args []string) error {
	imageSource, err := source.NewImageSource(src)
	if err != nil {
		return err
	}
	defer imageSource.Close()
	storage, err := storage.NewBoltStorage(src)
	if err != nil {
		return err
	}
	imgIndex, err := index.BuildInMemoryIndex(storage, 1)
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

	dstImg := imaging.New(aspectRatio.X*tileMultiple*tiles, aspectRatio.Y*tileMultiple*tiles, color.NRGBA{0, 0, 0, 0})
	progressBar := pb.StartNew(tiles * tiles)
	for i := 0; i < tiles; i++ {
		for j := 0; j < tiles; j++ {
			clip := imaging.Crop(referenceImg, image.Rectangle{
				Min: image.Point{X: i * aspectRatio.X, Y: j * aspectRatio.Y},
				Max: image.Point{X: (i + 1) * aspectRatio.X, Y: (j + 1) * aspectRatio.Y},
			})
			selected, err := imgIndex.Search(imaging.Resize(clip, aspectRatio.X, aspectRatio.Y, imaging.NearestNeighbor), aspectRatio)
			if err != nil {
				return err
			}
			selectedImg, err := imageSource.GetImage(selected)
			if err != nil {
				return err
			}
			dstImg = imaging.Paste(dstImg, imaging.Resize(selectedImg, aspectRatio.X*tileMultiple, 0, imaging.NearestNeighbor), image.Point{X: i * aspectRatio.X * tileMultiple, Y: j * aspectRatio.Y * tileMultiple})
			progressBar.Increment()
		}
	}
	progressBar.Finish()
	imaging.Save(dstImg, args[0]+".mosaic.jpg")
	return nil
}

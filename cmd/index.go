package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/timwu/mosaicer/analysis"
	"github.com/timwu/mosaicer/source"
	"github.com/timwu/mosaicer/storage"
)

var (
	indexCmd = &cobra.Command{
		Use:   "index",
		Short: "Index image files for photo mosaic generation",
		RunE:  doIndex,
	}
)

func init() {
	rootCmd.AddCommand(indexCmd)
}

func doIndex(cmd *cobra.Command, args []string) error {
	imageSource, err := source.NewImageSource(args[0])
	if err != nil {
		return err
	}
	defer imageSource.Close()
	names, err := imageSource.GetImageNames()
	if err != nil {
		return err
	}
	storage, err := storage.NewBoltStorage(args[0])
	if err != nil {
		return err
	}
	for i, name := range names {
		img, err := imageSource.GetImage(name)
		if err != nil {
			return err
		}
		data, err := analysis.Simple(img)
		if err != nil {
			return err
		}
		if err := storage.Store(name, data); err != nil {
			return err
		}
		log.Printf("Stored %s (%d/%d)", name, i+1, len(names))
	}
	return nil
}

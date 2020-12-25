package cmd

import (
	"log"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"github.com/timwu/mosaicer/analysis"
	"github.com/timwu/mosaicer/index"
	"github.com/timwu/mosaicer/source"
)

var (
	indexCmd = &cobra.Command{
		Use:   "index",
		Short: "Index image files for photo mosaic generation",
		RunE:  doIndex,
	}

	nThreads = 4
	samples  = 4
)

func init() {
	indexCmd.Flags().IntVar(&nThreads, "threads", 4, "Number of threads to use for indexing")
	indexCmd.Flags().IntVar(&samples, "samples", 4, "Number of samples per-image to take")
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
	boltIndex, err := index.NewBoltIndexBuilder(args[0])
	if err != nil {
		return err
	}
	defer boltIndex.Close()

	liveThreads := make(chan bool, nThreads)
	progressBar := pb.StartNew(len(names))
	for _, name := range names {
		liveThreads <- true
		name := name
		go func() {
			defer func() {
				<-liveThreads
				progressBar.Increment()
			}()
			img, err := imageSource.GetImage(name)
			if err != nil {
				log.Fatal(err)
			}
			data, err := analysis.Simple(img, samples)
			if err != nil {
				log.Fatal(err)
			}
			if err := boltIndex.Index(name, data); err != nil {
				log.Fatal(err)
			}
		}()
	}
	for i := 0; i < nThreads; i++ {
		liveThreads <- true
	}
	progressBar.Finish()
	return nil
}

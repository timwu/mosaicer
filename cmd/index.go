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
	"log"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"github.com/timwu/mosaicer/analysis"
	"github.com/timwu/mosaicer/index"
	"github.com/timwu/mosaicer/source"
	"github.com/timwu/mosaicer/util"
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

	limiter := util.NewLimiter(nThreads)
	progressBar := pb.StartNew(len(names))
	for _, name := range names {
		name := name
		limiter.Go(func() {
			defer progressBar.Increment()
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
		})
	}
	limiter.Close()
	progressBar.Finish()
	return nil
}

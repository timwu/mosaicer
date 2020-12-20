package main

import (
	"log"

	"github.com/timwu/mosaicer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

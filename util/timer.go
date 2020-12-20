package util

import (
	"log"
	"time"
)

// LogTime creates a simple time logger
func LogTime(s string) func() {
	startTime := time.Now()
	return func() {
		log.Printf("%s - %v", s, time.Now().Sub(startTime))
	}
}

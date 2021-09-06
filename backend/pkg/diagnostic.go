package pkg

import (
	"log"
	"time"
)

func Track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func Duration(msg string, start time.Time) {
	log.Printf("%v took: %v\n", msg, time.Since(start))
}

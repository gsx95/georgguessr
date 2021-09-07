package pkg

import (
	"log"
	"runtime"
	"time"
)

func Track() (string, time.Time) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function, time.Now()
}

func LogDuration(func_name string, start time.Time) {
	log.Printf("[DUR] %v  -  %v\n", time.Since(start), func_name)
}

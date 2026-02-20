package util

import (
	"fmt"
	"log"
	"os"
)

func EnableDebug() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Warnf prints a warning message to stderr
func Warnf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", v...)
}

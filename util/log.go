package util

import (
	"log"
	"os"
)

func EnableDebug() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

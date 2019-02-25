package main

import (
	"log"
	"os"
	"time"

	"github.com/udhos/inkey/inkey"
)

func main() {
	input := inkey.New(os.Stdin)

	log.Printf("type 'q' to quit")

	for {
		b, found := input.Inkey()
		if found {
			log.Printf("key found: '%c' value=%d", b, b)
			if b == 'q' {
				break
			}
			continue
		}
		log.Printf("waiting for key")
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("done")
}

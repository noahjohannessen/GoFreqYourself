package main

import (
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
)

func main() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatalf("PortAudio initialization failed: %v", err)
	}
	defer portaudio.Terminate()

	fmt.Println("PortAudio is working!")
}

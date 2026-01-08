package main

import (
	"elp-project/internal/audio"
	"fmt"
)

func main() {
	_, wd, err := audio.ParseWav("assets/sine_8k.wav")
	if err != nil {
		panic(err)
	}

	var step int = 2048
	for {
		chunk, eof := wd.Advance(step)
		if len(chunk) > 0 {
			fmt.Println("Chunk size (bytes): ", len(chunk))
			fmt.Println("ChunkID : ", wd.ChunkID)
			// fmt.Printf("%x\n", wd.Samples)
		}
		if eof {
			break
		}
	}
}

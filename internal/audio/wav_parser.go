// Read WAV file

package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func ParseWav(path string) (WavType, error) {
	// f, err := os.Open("assets/sample-3s.wav")
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Can't open file: ", err)
	}
	defer f.Close()

	// Get RIFF chunk + chunk describing data format
	var header WavHeader
	err = binary.Read(f, binary.LittleEndian, &header)
	if err != nil {
		panic(err)
	}
	header.logHeader()

	fmt.Printf("\n")

	var dataInfo WavDataChunk
	err = dataInfo.FindDataChunk(f)
	if err != nil {
		log.Fatal(err)
	}
	dataInfo.logDataChunk()

	return WavType{header, dataInfo}, nil
}

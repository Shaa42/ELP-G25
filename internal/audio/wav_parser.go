// Read WAV file

package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func ParseWav(path string) (WavType, WavData, error) {
	// Open WAV file
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Get RIFF chunk + chunk describing data format
	var header WavHeader
	err = binary.Read(f, binary.LittleEndian, &header)
	if err != nil {
		return WavType{}, WavData{}, err
	}
	header.logHeader()

	fmt.Printf("\n")

	// Get to data chunk
	var dataInfo WavDataChunk
	err = dataInfo.FindDataChunk(f)
	if err != nil {
		return WavType{}, WavData{}, err
	}
	dataInfo.logDataChunk()

	// Read audio samples
	samples := make([]byte, dataInfo.DataSize)
	_, err = io.ReadFull(f, samples)
	if err != nil {
		return WavType{}, WavData{}, err
	}

	// Build WavData with metadata & cursor
	wd := WavData{
		Metadata: WavMetadata{
			SampleRate: header.Frequency,
			Channels:   header.NbrChannels,
			Bitdepth:   header.BitsPerSample,
			Format:     header.AudioFormat,
		},
		Samples:       samples,
		ChunkID:       0,
		cursorSamples: 0,
	}

	return WavType{header, dataInfo}, wd, nil
}

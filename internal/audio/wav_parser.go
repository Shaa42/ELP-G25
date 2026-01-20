// Read WAV file

package audio

import (
	"encoding/binary"
	"io"
	"os"
)

func ParseWav(path string) (WavType, *WavData, error) {
	// Open WAV file
	f, err := os.Open(path)
	if err != nil {
		return WavType{}, nil, err
	}

	// Get RIFF chunk + chunk describing data format
	var header WavHeaderInfo
	err = binary.Read(f, binary.LittleEndian, &header)
	if err != nil {
		f.Close()
		return WavType{}, nil, err
	}
	// header.logHeader()

	// fmt.Printf("\n")

	// Get to data chunk
	var dataInfo WavDataInfo
	err = dataInfo.FindDataChunk(f)
	if err != nil {
		f.Close()
		return WavType{}, nil, err
	}
	// dataInfo.logDataChunk()

	// Keep in memory audio data cursor initial position
	dataOffset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		f.Close()
		return WavType{}, nil, err
	}

	// Compute frame size and total frames number
	frameSize := int(header.NbrChannels) * int(header.BitsPerSample/8)
	totalFrames := int(dataInfo.DataSize) / frameSize

	// Build WavData with metadata & cursor
	wd := &WavData{
		Metadata: WavMetadata{
			SampleRate: header.Frequency,
			Channels:   header.NbrChannels,
			Bitdepth:   header.BitsPerSample,
			Format:     header.AudioFormat,
		},
		File:          f,
		DataOffset:    dataOffset,
		DataSize:      dataInfo.DataSize,
		ChunkID:       0,
		CursorSamples: 0,
		TotalFrames:   totalFrames,
	}

	return WavType{header, dataInfo}, wd, nil
}

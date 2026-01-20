package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

/* ####################################################################### */

type WavData struct {
	Metadata      WavMetadata
	File          *os.File
	DataOffset    int64
	DataSize      uint32
	ChunkID       int
	CursorSamples int
	TotalFrames   int
}

func (wd *WavData) Advance(n int) (chunk []byte, eof bool) {
	if n <= 0 {
		return nil, wd.isEOF()
	}

	frameSize := wd.bytesPerFrame()

	if wd.CursorSamples >= wd.TotalFrames {
		return nil, true
	}

	// Compute how much frames there is to read
	framesToRead := n
	if wd.CursorSamples+framesToRead > wd.TotalFrames {
		framesToRead = wd.TotalFrames - wd.CursorSamples
	}

	// Create data buffer
	bytesToRead := framesToRead * frameSize
	chunk = make([]byte, bytesToRead)

	// Read from file
	bytesRead, err := wd.File.Read(chunk)
	if err != nil && err != io.EOF {
		return nil, true
	}

	// Update file cursor
	framesRead := bytesRead / frameSize
	wd.CursorSamples += framesRead
	wd.ChunkID++

	eof = wd.CursorSamples >= wd.TotalFrames
	return chunk, eof
}

// Reset file cursor
func (wd *WavData) Reset() error {
	_, err := wd.File.Seek(wd.DataOffset, io.SeekStart)
	if err != nil {
		return err
	}
	wd.CursorSamples = 0
	wd.ChunkID = 0
	return nil
}

// Close wav file
func (wd *WavData) Close() error {
	if wd.File != nil {
		return wd.File.Close()
	}
	return nil
}

// Return how many samples are left
func (wd *WavData) RemainingSamples() int {
	if wd.CursorSamples >= wd.TotalFrames {
		return 0
	}
	return wd.TotalFrames - wd.CursorSamples
}

func (wd *WavData) isEOF() bool {
	return wd.CursorSamples >= wd.TotalFrames
}

func (wd *WavData) bytesPerFrame() int {
	return int(wd.Metadata.Channels) * int(wd.Metadata.Bitdepth/8)
}

/* ####################################################################### */

type WavDataChunk struct {
	Metadata WavMetadata
	ChunkID  int
	Samples  []byte
}

func (wdc WavDataChunk) Len() int {
	return len(wdc.Samples)
}

func (wdc WavDataChunk) ConvSampByteToFloat32() []float32 {
	bytesPerSample := int(wdc.Metadata.Bitdepth / 8)
	numSamples := wdc.Len() / bytesPerSample
	samples := make([]float32, numSamples)

	switch wdc.Metadata.Bitdepth {
	case 8:
		for i := range numSamples {
			val := int(wdc.Samples[i]) - 128
			samples[i] = float32(val) / 128.0
		}

	case 16:
		for i := range numSamples {
			val := int16(binary.LittleEndian.Uint16(wdc.Samples[i*2 : i*2+2]))
			samples[i] = float32(val) / 32768.0
		}
	}

	return samples
}

func (wdc *WavDataChunk) ConvSampFloat32ToByte(samples []float32) {
	switch wdc.Metadata.Bitdepth {
	case 8:
		// 8 bits: UNSIGNED (0-255)
		// -1.0 → 0, 0.0 → 128, 1.0 → 255
		for i, sample := range samples {
			// Clipping
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}

			// Convertir [-1.0, 1.0] → [-128, 127] → [0, 255]
			val := int(sample * 128.0)
			if val > 127 {
				val = 127
			} else if val < -128 {
				val = -128
			}
			wdc.Samples[i] = byte(val + 128)
		}

	case 16:
		// 16 bits: SIGNED (-32768 à 32767)
		// -1.0 → -32768, 0.0 → 0, 1.0 → 32767
		for i, sample := range samples {
			// Clipping
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}

			// Convertir [-1.0, 1.0] → [-32768, 32767]
			val := int32(sample * 32768.0)
			if val > 32767 {
				val = 32767
			} else if val < -32768 {
				val = -32768
			}

			// Écrire en little-endian (2 bytes)
			offset := i * 2
			binary.LittleEndian.PutUint16(wdc.Samples[offset:offset+2], uint16(val))
		}
	}
}

/* ####################################################################### */

type WavDataInfo struct {
	DataBlocID [4]byte
	DataSize   uint32
}

func (data WavDataInfo) logDataChunk() {
	fmt.Println("=== WAV DATA CHUNK ===")
	fmt.Printf("DataBlocID : %s\n", data.DataBlocID)
	fmt.Printf("DataSize : %d\n", data.DataSize)
}

func (dataChunk *WavDataInfo) FindDataChunk(f *os.File) error {
	for {
		var chunkID [4]byte
		err := binary.Read(f, binary.LittleEndian, &chunkID)
		if err != nil {
			return err
		}

		var chunkSize uint32
		err = binary.Read(f, binary.LittleEndian, &chunkSize)
		if err != nil {
			return err
		}

		id := string(chunkID[:])
		if id == "data" {
			copy(dataChunk.DataBlocID[:], chunkID[:])
			dataChunk.DataSize = chunkSize
			return nil
		}

		// Sauter le chunk non-data
		_, err = f.Seek(int64(chunkSize), io.SeekCurrent)
		if err != nil {
			return err
		}

		// Ajouter 1 si chunkSize impair
		if chunkSize%2 == 1 {
			_, err = f.Seek(1, io.SeekCurrent)
			if err != nil {
				return err
			}
		}
	}
}

/* ####################################################################### */

type WavHeaderInfo struct {
	FileTypeBlocID [4]byte
	FileSize       uint32
	FileFormatID   [4]byte

	FormatBlocID  [4]byte
	BlocSize      uint32
	AudioFormat   uint16
	NbrChannels   uint16
	Frequency     uint32
	BytePerSec    uint32
	BytePerBloc   uint16
	BitsPerSample uint16
}

func (header WavHeaderInfo) logHeader() {
	fmt.Println("=== WAV HEADER ===")
	fmt.Printf("FileTypeBlocID : %s\n", header.FileTypeBlocID)
	fmt.Printf("FileSize       : %d\n", header.FileSize)
	fmt.Printf("FileFormatID   : %s\n", header.FileFormatID)

	fmt.Printf("FormatBlocID   : %s\n", header.FormatBlocID)
	fmt.Printf("BlocSize       : %d\n", header.BlocSize)
	fmt.Printf("AudioFormat    : %d\n", header.AudioFormat)
	fmt.Printf("NbrChannels    : %d\n", header.NbrChannels)
	fmt.Printf("Frequency      : %d Hz\n", header.Frequency)
	fmt.Printf("BytePerSec     : %d\n", header.BytePerSec)
	fmt.Printf("BytePerBloc    : %d\n", header.BytePerBloc)
	fmt.Printf("BitsPerSample  : %d\n", header.BitsPerSample)
}

/* ####################################################################### */

type WavMetadata struct {
	SampleRate uint32
	Channels   uint16
	Bitdepth   uint16
	Format     uint16
}

/* ####################################################################### */

type WavType struct {
	WavHeader    WavHeaderInfo
	WavDataChunk WavDataInfo
}

func (wt WavType) Log() {
	wt.WavHeader.logHeader()
	wt.WavDataChunk.logDataChunk()
}

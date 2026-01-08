package audio

import "fmt"

type WavHeader struct {
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

func (header WavHeader) logHeader() {
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

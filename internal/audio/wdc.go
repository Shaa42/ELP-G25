package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type WavDataChunk struct {
	DataBlocID [4]byte
	DataSize   uint32
}

func (data WavDataChunk) logDataChunk() {
	fmt.Println("=== WAV DATA CHUNK ===")
	fmt.Printf("DataBlocID : %s\n", data.DataBlocID)
	fmt.Printf("DataSize : %d\n", data.DataSize)
}

func (dataChunk *WavDataChunk) FindDataChunk(f *os.File) error {
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

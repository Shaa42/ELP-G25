package main

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"elp-project/internal/audio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <wav_path> [server_addr]\n", os.Args[0])
		os.Exit(1)
	}
	wavPath := os.Args[1]
	addr := "localhost:42069"
	if len(os.Args) >= 3 {
		addr = os.Args[2]
	}

	// Construire WavData à partir du fichier WAV
	wd, err := buildWavData(wavPath)
	if err != nil {
		log.Fatalf("failed to build WavData: %v", err)
	}

	// Connexion au serveur
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("failed to dial %s: %v", addr, err)
	}
	defer conn.Close()

	// Envoyer WavData en plusieurs chunks via gob et lire un ACK pour chaque chunk
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	frameSize := int(wd.Metadata.Channels) * int(wd.Metadata.Bitdepth/8)
	if frameSize <= 0 {
		log.Fatalf("invalid frame size computed from metadata: channels=%d bitdepth=%d", wd.Metadata.Channels, wd.Metadata.Bitdepth)
	}
	totalFrames := len(wd.Samples) / frameSize
	chunkFrames := 2048
	chunkID := 0

	for startFrame := 0; startFrame < totalFrames; startFrame += chunkFrames {
		endFrame := min(startFrame+chunkFrames, totalFrames)

		startByte := startFrame * frameSize
		endByte := endFrame * frameSize

		msg := audio.WavData{
			Metadata: wd.Metadata,
			Samples:  wd.Samples[startByte:endByte],
			ChunkID:  chunkID,
		}

		if err := enc.Encode(msg); err != nil {
			log.Fatalf("failed to gob-encode WavData chunk %d: %v", chunkID, err)
		}

		var ack string
		if err := dec.Decode(&ack); err != nil {
			log.Fatalf("failed to gob-decode ACK for chunk %d: %v", chunkID, err)
		}
		fmt.Println("Client: reçu:", ack)

		chunkID++
	}

	fmt.Printf("Client: envoyé %d chunks (%d frames au total) au serveur %s\n", chunkID, totalFrames, addr)
}

func buildWavData(path string) (audio.WavData, error) {
	f, err := os.Open(path)
	if err != nil {
		return audio.WavData{}, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	// Lire l'entête WAV (même struct que côté package audio)
	var header audio.WavHeader
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return audio.WavData{}, fmt.Errorf("read header: %w", err)
	}

	// Atteindre le chunk data
	var dataInfo audio.WavDataChunk
	if err := dataInfo.FindDataChunk(f); err != nil {
		return audio.WavData{}, fmt.Errorf("find data chunk: %w", err)
	}

	// Lire les samples bruts
	samples := make([]byte, dataInfo.DataSize)
	if _, err := io.ReadFull(f, samples); err != nil {
		return audio.WavData{}, fmt.Errorf("read samples: %w", err)
	}

	// Construire WavData
	wd := audio.WavData{
		Metadata: audio.WavMetadata{
			SampleRate: header.Frequency,
			Channels:   header.NbrChannels,
			Bitdepth:   header.BitsPerSample,
			Format:     header.AudioFormat,
		},
		Samples: samples,
		ChunkID: 0,
	}
	return wd, nil
}

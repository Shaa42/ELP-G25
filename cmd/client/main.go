package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"

	"elp-project/internal/audio"
)

func main() {
	wavPath := "assets/sample-3s.wav"
	addr := "localhost:42069"

	// Construire WavData à partir du fichier WAV
	wt, wd, err := audio.ParseWav(wavPath)
	if err != nil {
		log.Printf("failed to build WavData: %v", err)
		panic(err)
	}
	defer wd.Close()

	wt.Log()

	chunkSize := 4096
	totalChunks := (wd.TotalFrames + chunkSize - 1) / chunkSize

	// Connexion au serveur
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("failed to dial %s: %v", addr, err)
	}
	defer conn.Close()

	// Chunks chan
	// serverChunks := make(chan audio.WavDataChunk, totalChunks)

	receivedChunks := make([]audio.WavDataChunk, totalChunks)
	var mut sync.Mutex

	// Waitgroup
	var wg sync.WaitGroup

	// Envoyer WavData en plusieurs chunks via gob et lire un ACK pour chaque chunk
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	wg.Go(func() {
		for range totalChunks {
			// Réception du chunk traité par le serveur
			var chunk audio.WavDataChunk
			if err := dec.Decode(&chunk); err != nil {
				log.Printf("Error receiving chunk: %v", err)
				break
			}

			// Debug
			fmt.Printf("Received chunk #%d (%d bytes)\n", chunk.ChunkID, len(chunk.Samples))

			// Enregistre dans une liste
			mut.Lock()
			if chunk.ChunkID > 0 && chunk.ChunkID <= totalChunks {
				receivedChunks[chunk.ChunkID-1] = chunk
			}
			mut.Unlock()
		}
		fmt.Println("done reception")
	})

	// Iter through data until EOF
	for {
		chunk, eof := wd.Advance(chunkSize)
		if len(chunk) > 0 {
			wdc := audio.WavDataChunk{
				Metadata: wd.Metadata,
				ChunkID:  wd.ChunkID,
				Samples:  chunk,
			}

			// Encode Struct to send to server
			if err := enc.Encode(wdc); err != nil {
				log.Fatalf("failed to gob-encode WavData chunk %d: %v", wd.ChunkID, err)
			}
		}

		if eof {
			fmt.Println("EOF")
			break
		}
	}

	wg.Wait()

	count := 0
	for _, c := range receivedChunks {
		if len(c.Samples) > 0 {
			count++
		}
	}

	if count != totalChunks {
		log.Printf("chunks received (%d) != chunks sent (%d)\n",
			count, totalChunks)
	}

	fmt.Println("Creating new file in", "assets/output.wav")
	audio.WriteWav("assets/output.wav", receivedChunks)
}

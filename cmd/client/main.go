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
	wavPath := "assets/sine_8k.wav"
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
	serverChunks := make(chan audio.WavDataChunk, totalChunks)

	// Waitgroup
	var wg sync.WaitGroup

	// Envoyer WavData en plusieurs chunks via gob et lire un ACK pour chaque chunk
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(serverChunks)

		for range totalChunks {
			var chunk audio.WavDataChunk
			if err := dec.Decode(&chunk); err != nil {
				log.Printf("Erreur réception chunk: %v", err)
				break
			}
			fmt.Printf("Reçu chunk #%d (%d bytes)\n", chunk.ChunkID, len(chunk.Samples))
			serverChunks <- chunk
		}
		fmt.Println("Réception terminée")
	}()

	// Iter through data until EOF
	for {
		chunk, eof := wd.Advance(chunkSize)
		if len(chunk) > 0 {
			wdc := audio.WavDataChunk{
				Metadata: wd.Metadata,
				ChunkID:  wd.ChunkID,
				Samples:  chunk,
			}
			// fmt.Printf("Chunk #%d: %d bytes\n", wd.ChunkID, wdc.Len())

			// Encode Struct to send to server
			if err := enc.Encode(wdc); err != nil {
				log.Fatalf("failed to gob-encode WavData chunk %d: %v", wd.ChunkID, err)
			}

			// Get server ACK message
			// var ack string
			// if err := dec.Decode(&ack); err != nil {
			// 	log.Fatalf("failed to gob-decode ACK for chunk %d: %v", wd.ChunkID, err)
			// }
			// fmt.Println("Client: reçu:", ack)

			// fmt.Printf("Client: envoyé %d chunks (%d frames au total) au serveur %s\n", wd.ChunkID, wd.TotalFrames, addr)
		}

		if eof {
			fmt.Println("EOF")
			break
		}
	}
	wg.Wait()
}

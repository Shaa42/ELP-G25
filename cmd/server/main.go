// Server: receive WavData via gob and reply with per-chunk ACK
package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"

	"elp-project/internal/audio"
	"elp-project/internal/processor"
)

// Handle the connection with the client
func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Server: new connection from", conn.RemoteAddr())

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	// Wait for the client to send over data, process it and wait for the client to close the conn
	for {
		var wdc audio.WavDataChunk
		if err := dec.Decode(&wdc); err != nil {
			if err == io.EOF {
				log.Println("Server: client closed connection", conn.RemoteAddr())
				return
			}
			log.Println("Server: gob decode error:", err)
			return
		}

		// ack := fmt.Sprintf("ACK Chunk %d", wdc.ChunkID)
		// if err := enc.Encode(ack); err != nil {
		// 	log.Println("Server: gob encode ACK error:", err)
		// 	return
		// }

		HandleSample(&wdc)

		if err := enc.Encode(wdc); err != nil {
			log.Println("Server: gob encode ACK error:", err)
			return
		}

		// log.Printf("Server: processed WavData (Samples=%d, SR=%d, Ch=%d, Bits=%d) -> %q\n",
		// wdc.Len(), wdc.Metadata.SampleRate, wdc.Metadata.Channels, wdc.Metadata.Bitdepth, ack)
	}
}

func HandleSample(wdc *audio.WavDataChunk) {
	var samplesFloat32 []float32
	samplesFloat32 = wdc.ConvSampByteToFloat32()
	samplesFloat32 = processor.AddDB(samplesFloat32, 2.0)

	wdc.ConvSampFloat32ToByte(samplesFloat32)
}

func main() {
	// Open TCP listening server
	addr := ":42069"
	log.Println("Server: listening on", addr)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Server: failed to listen:", err)
	}
	defer ln.Close()

	// Wait for incoming connection request and create new process to handle the new connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Server: accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

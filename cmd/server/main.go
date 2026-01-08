// Server: receive WavData via gob and reply with per-chunk ACK
package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"

	"elp-project/internal/audio"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Server: new connection from", conn.RemoteAddr())

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	for {
		var wd audio.WavData
		if err := dec.Decode(&wd); err != nil {
			if err == io.EOF {
				log.Println("Server: client closed connection", conn.RemoteAddr())
				return
			}
			log.Println("Server: gob decode error:", err)
			return
		}

		ack := fmt.Sprintf("ACK Chunk %d", wd.ChunkID)
		if err := enc.Encode(ack); err != nil {
			log.Println("Server: gob encode ACK error:", err)
			return
		}

		log.Printf("Server: processed WavData (Samples=%d, SR=%d, Ch=%d, Bits=%d) -> %q\n",
			len(wd.Samples), wd.Metadata.SampleRate, wd.Metadata.Channels, wd.Metadata.Bitdepth, ack)
	}
}

func main() {
	addr := ":42069"
	log.Println("Server: listening on", addr)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Server: failed to listen:", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Server: accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

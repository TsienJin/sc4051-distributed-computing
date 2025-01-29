package server

import (
	"fmt"
	"log"
	"net"
	"server/internal/env"
	"server/internal/handle"
	"server/internal/pools"
	"server/internal/protocol/proto_defs"
)

func Serve() {

	staticEnv := env.GetStaticEnv()

	// Determine server's address on given port
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", staticEnv.ServerPort))
	if err != nil {
		log.Fatalln("Error resolving server address: ", err)
		return
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalln("Error starting server: ", err)
		return
	}
	defer conn.Close()
	log.Printf("UDP Server listening on %s\n", conn.LocalAddr().String())

	// Reading packets
	readBuffer := make([]byte, proto_defs.PacketSizeLimit)

	for {
		n, addr, err := conn.ReadFromUDP(readBuffer)
		if err != nil {
			log.Println("Error reading into buffer: ", err)
			continue
		}

		dataBuf := pools.PacketBytesPool.Get().([]byte)
		copy(dataBuf, readBuffer[:n])

		go handle.IncomingPacket(conn, *addr, n, dataBuf)

		log.Printf("Received %d bytes from %v\n", n, addr)
	}

}

package server

import (
	"fmt"
	"log/slog"
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
		slog.Error("Error resolving server address: ", "err", err)
		return
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		slog.Error("Error starting server: ", "err", err)
		return
	}
	defer conn.Close()
	slog.Info(fmt.Sprintf("UDP Server listening on %s\n", conn.LocalAddr().String()))

	// Reading packets
	readBuffer := make([]byte, proto_defs.PacketSizeLimit)

	for {
		n, addr, err := conn.ReadFromUDP(readBuffer)
		if err != nil {
			slog.Error("Error reading into buffer: ", "err", err)
			continue
		}

		dataBuf := pools.PacketBytesPool.Get().([]byte)
		copy(dataBuf, readBuffer[:n])

		go handle.IncomingPacket(conn, *addr, n, dataBuf)

		slog.Info(fmt.Sprintf("Received %d bytes from %v\n", n, addr))
	}

}

package server

import (
	"fmt"
	"log/slog"
	"net"
	"server/internal/handle"
	"server/internal/pools"
	"server/internal/protocol/proto_defs"
)

func serveOnConn(conn *net.UDPConn) {

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

		// Each incoming packet is spun onto its own GoRoutine; therefore for each packet, no other
		// GoRoutine needs to be initiated.
		go handle.IncomingPacket(conn, addr, n, dataBuf)

		slog.Info(fmt.Sprintf("Received %d bytes from %v\n", n, addr))
	}

}

func Serve(port int) {

	// Determine server's address on given port
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
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
	serveOnConn(conn)

}

func ServeRandomPort() (int, error) {
	// Resolve a UDP address with port 0 (OS assigns a free port)
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return 0, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Listen on a random available port
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen on a UDP port: %w", err)
	}

	// Extract the assigned port
	port := conn.LocalAddr().(*net.UDPAddr).Port
	slog.Info(fmt.Sprintf("Assigned random UDP port: %d", port))

	// Start serving packets
	go serveOnConn(conn)

	return port, nil
}

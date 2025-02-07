package logging

import (
	"fmt"
	"log/slog"
	"net"
)

func Serve(port int) {

	// Create server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("Unable to create logging server", "err", err)
		panic(err)
	}

	handler := GetConnectionHandler()
	defer handler.CloseAndRemoveClients()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				slog.Error("Unable to accept new logging connection", "err", err)
				continue
			}
			slog.Info("Accepted new logging connection")
			go HandleClientMessages(conn.(*net.TCPConn))
			handler.AddClient(conn.(*net.TCPConn))
		}
	}()

	for m := range GetMessageQueue() {
		handler.SendMessage(m)
		SendToMatterMost(m)
	}

}

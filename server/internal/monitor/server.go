package monitor

import (
	"fmt"
	"log/slog"
	"net"
	"server/internal/vars"
)

func Serve() {

	// Create server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", vars.GetStaticEnv().ServerLogPort))
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
			handler.AddClient(conn.(*net.TCPConn))
		}
	}()

	for m := range GetMessageQueue() {
		handler.SendMessage(m)
	}

}

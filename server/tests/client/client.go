package client

import (
	"context"
	"log/slog"
	"net"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/response"
	"sync"
)

type Client struct {
	wg sync.WaitGroup

	logger *slog.Logger
	name   string

	manager      *sendManager
	targetServer *net.UDPAddr
	conn         *net.UDPConn

	responseBytes chan [proto_defs.PacketSizeLimit]byte // This chan is used internally for message passing
	Responses     chan *response.Response               // Exposed to process incoming messages

	Ctx    context.Context
	Cancel context.CancelFunc
}

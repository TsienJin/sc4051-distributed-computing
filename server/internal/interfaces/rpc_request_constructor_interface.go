package interfaces

import "server/internal/protocol"

type RpcRequestConstructor func() ([]*protocol.Packet, error)

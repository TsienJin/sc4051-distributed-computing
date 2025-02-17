package request_constructor

import (
	"server/internal/interfaces"
	"server/internal/protocol"
)

// NewDoNothingPlaceholder is a constructor meant to allow the SyncValidator to only receive packet without sending.
func NewDoNothingPlaceholder() interfaces.RpcRequestConstructor {
	return func() ([]*protocol.Packet, error) {
		return nil, nil
	}
}

package pools

import (
	"server/internal/protocol/proto_defs"
	"sync"
)

var PacketBytesPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, proto_defs.PacketSizeLimit)
	},
}

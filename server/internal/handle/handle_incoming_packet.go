package handle

import (
	"net"
	"server/internal/chance"
	"server/internal/pools"
)

func IncomingPacket(
	conn *net.UDPConn,
	addr net.UDPAddr,
	nBytes int,
	data []byte,
) {
	defer pools.PacketBytesPool.Put(data)

	// Drop Chance
	if chance.DropPacket() {
		return
	}

	// Validate checksum

}

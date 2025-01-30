package handle

import (
	"fmt"
	"log/slog"
	"net"
	"server/internal/chance"
	"server/internal/pools"
	"server/internal/protocol"
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
		slog.Warn(fmt.Sprintf("[IN:DROP] %d from %s", nBytes, addr.String()))
		return
	}

	// Validate checksum
	if !protocol.ValidateChecksumBytes(data[:nBytes-4], data[nBytes-4:nBytes]) {
		slog.Error(fmt.Sprintf("[IN:CHECKSUM] %d from %s failed checksum check", nBytes, addr.String()))
		return
	}

	// Unmarshall packet

}

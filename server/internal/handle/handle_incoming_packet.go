package handle

import (
	"fmt"
	"log/slog"
	"net"
	"server/internal/chance"
	"server/internal/network"
	"server/internal/pools"
	"server/internal/protocol"
	"server/internal/protocol/constructors"
)

func IncomingPacket(
	conn *net.UDPConn,
	addr *net.UDPAddr,
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
	// We have to reference nBytes here, since the pools.PacketBytesPool must contain [MaxSize]byte
	if !protocol.ValidateChecksumBytes(data[:nBytes-4], data[nBytes-4:nBytes]) {
		slog.Error(fmt.Sprintf("[IN:CHECKSUM] %d from %s failed checksum check", nBytes, addr.String()))
		return
	}

	// Unmarshall packet
	var packet protocol.Packet
	err := packet.UnmarshalBinary(data)
	if err != nil {
		slog.Error("[IN:UNMARSHAL] Unable to unmarshal packet from binary")
		return
	}

	// Handle acknowledgements
	if packet.Header.Flags.AckRequired() {
		ackPacket, err := constructors.NewAck(packet.Header.MessageId, packet.Header.PacketNumber)
		if err != nil {
			slog.Error("[IN:ACK] Unable to create ack packet to be sent")
		}
		if err := network.SendPacket(conn, addr, ackPacket); err != nil {
			slog.Error("Unable to send ack packet", "AckPacket", ackPacket)
		}
	}

	// Pass off to message assembly
	slog.Info("[IN:HANDOFF] Packet validated and acknowledged, handing off to assembler")
	AssembleMessageFromPacket(conn, addr, &packet)

}

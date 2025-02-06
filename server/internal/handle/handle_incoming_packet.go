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
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/response"
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
		slog.Info("[IN:ACK] Packet acknowledged")
	}

	switch packet.Header.MessageType {
	case proto_defs.MessageTypeAcknowledge:
		slog.Info("[IN:SORT] Sent packet acknowledged, removing from history")
		var ackPayload protocol.AckResendPayload
		if err := ackPayload.UnmarshalBinary(packet.Payload); err != nil {
			slog.Error("Unable to unmarshal ack payload", "err", err)
		}
		ident := ackPayload.ToPacketIdent()
		// Packet has been confirmed to be received
		network.GetSendHistoryInstance().Remove(*ident)
		// Once first ack of res packet is received, the response is removed
		response.GetResponseHistoryInstance().RemoveResponse(ident.MessageId)
		break
	case proto_defs.MessageTypeRequestResend:
		slog.Info("[IN:SORT] Requesting for packet resend")
		RequestResendPacket(conn, addr, &packet)
		break
	default:
		// Pass off to message assembly
		slog.Info("[IN:HANDOFF] Packet validated and acknowledged, handing off to assembler")
		AssembleMessageFromPacket(conn, addr, &packet)
		break
	}

}

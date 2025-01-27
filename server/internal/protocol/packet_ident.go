package protocol

import "server/internal/protocol/proto_defs"

type PacketIdent struct {
	MessageId    proto_defs.MessageId
	PacketNumber uint16
}

func ExtractIdentFromPacket(p *Packet) PacketIdent {
	return PacketIdent{
		MessageId:    p.Header.MessageId,
		PacketNumber: p.Header.PacketNumber,
	}
}

func ExtractIdentFromPacketHeader(p *PacketHeader) PacketIdent {
	return PacketIdent{
		MessageId:    p.MessageId,
		PacketNumber: p.PacketNumber,
	}
}

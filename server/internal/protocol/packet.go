package protocol

// Packet s are the data structs used to represent the underlying data.
// there is a limit size of proto_def.PacketSizeLimit
type Packet struct {
	Header  PacketHeader
	Payload []byte
}

func NewPacket(h PacketHeader, p []byte) *Packet {
	return &Packet{
		Header:  h,
		Payload: p,
	}
}

func (p *Packet) ToBytes() ([]byte, error) {

	return []byte{}, nil
}

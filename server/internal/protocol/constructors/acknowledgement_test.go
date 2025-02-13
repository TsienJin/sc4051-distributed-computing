package constructors

import (
	"github.com/google/go-cmp/cmp"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"testing"
)

func TestNewAck(t *testing.T) {

	p, err := NewAck(proto_defs.NewMessageId(), 0)
	if err != nil {
		t.Error(err)
	}

	bin, err := p.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}

	var packet protocol.Packet
	if err := packet.UnmarshalBinary(bin[:]); err != nil {
		t.Error(err)
		return
	}

	if !cmp.Equal(*p, packet) {
		t.Logf("%v\n", p)
		t.Logf("%v\n", packet)
		t.Error("Reflected ack packet does not match!")
		return
	}

}

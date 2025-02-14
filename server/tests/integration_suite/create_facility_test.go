package integration_suite

import (
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"testing"
	"time"
)

func TestCreateFacility_successful(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestCreateFacility_successful"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(5)*time.Second),
	)
	defer c.Close()

	payload := &request.FacilityCreatePayload{
		Name: bookings.FacilityName("TestCreateFacility_successful"),
	}

	payloadBytes, err := payload.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	r := request.Request{
		MethodIdentifier: request.MethodIdentifierFacilityCreate,
		Payload:          payloadBytes,
	}

	headerDistilled := &protocol.PacketHeaderDistilled{
		Version:     proto_defs.ProtocolV1,
		MessageId:   proto_defs.NewMessageId(),
		MessageType: proto_defs.MessageTypeRequest,
		RequireAck:  true,
	}

	message, err := protocol.NewMessage(headerDistilled, &r)
	if err != nil {
		t.Error(err)
	}

	packets, err := message.ToPackets()
	if err != nil {
		t.Error(err)
	}

	ok := false

	if err := c.SendPackets(packets); err != nil {
		t.Error(err)
	}

LOOP:
	for {
		select {
		case <-c.Ctx.Done():
			break LOOP
		case r := <-c.Responses: // we should only expect 1 response (ok or error)

			var res response.Response
			if err := res.UnmarshalBinary(r.Payload); err != nil {
				t.Error(err)
				return
			}
			ok = res.StatusCode == response.StatusOk
			break LOOP
		default:
			continue
		}
	}

	if !ok {
		t.Error("Test did not pass, check if response was received")
	}

}

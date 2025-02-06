package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func FacilityQuery(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Unmarshal into payload
	var p request.FacilityQueryPayload
	if err := p.UnmarshalBinary(message.Payload); err != nil {
		slog.Error("Unable to unmarshal FacilityQueryPayload", "err", err)
		return
	}

	// Query facility
	m := bookings.GetManager()
	r, err := m.QueryFacility(p.Name, p.Days)
	if err != nil {
		slog.Error("Unable to execute query", "FacilityName", p.Name, "Days", p.Days, "err", err)
		response.SendResponse(c, a, response.NewResponse(
			response.WithOriginalMessageId(message.Header.MessageId),
			response.WithStatusCode(response.StatusBadRequest),
		))
	}
	slog.Info("Successfully queried facility", "FacilityName", p.Name, "Days", p.Days, "Res", r)
	response.SendResponse(c, a, response.NewResponse(
		response.WithOriginalMessageId(message.Header.MessageId),
		response.WithStatusCode(response.StatusOk),
		response.WithPayloadBytes(r),
	))
}

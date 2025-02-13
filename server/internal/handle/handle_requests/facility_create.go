package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func FacilityCreate(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get message payload unmarshalled
	var p request.FacilityCreatePayload
	if err := p.UnmarshalBinary(message.Payload[1:]); err != nil {
		slog.Error("Unable to unmarshall FacilityCreatePayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	// Create facility
	m := bookings.GetManager()
	err := m.NewFacility(p.Name)
	if err != nil {
		slog.Error("Unable to create new Facility", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	// Facility successfully created
	slog.Info("Successfully created facility", "Facility", p.Name)
	response.SendResponse(c, a, response.NewOkResponse(message.Header.MessageId))
}

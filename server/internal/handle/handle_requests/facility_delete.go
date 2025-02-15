package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func FacilityDelete(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get payload
	var p request.FacilityDeletePayload
	if err := p.UnmarshalBinary(message.Payload[1:]); err != nil {
		slog.Error("Unable to unmarshall FacilityDeletePayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	// Process the request
	m := bookings.GetManager()
	err := m.DeleteFacility(p.Name)
	if err != nil {
		slog.Error("Unable to delete Facility", "FacilityName", p.Name, "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	// Successfully deleted
	slog.Info("Successfully deleted facility, sending response", "FacilityName", p.Name)
	response.SendResponse(c, a, response.NewOkResponse(message.Header.MessageId))

}

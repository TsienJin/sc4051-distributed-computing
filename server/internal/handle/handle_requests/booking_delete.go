package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func BookingDelete(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get message payload unmarshalled
	var p request.BookingDeletePayload
	if err := p.UnmarshalBinary(message.Payload); err != nil {
		slog.Error("Unable to unmarshall BookingDeletePayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	// Delete facility
	m := bookings.GetManager()
	err := m.DeleteBookingFromId(p.Id)
	if err != nil {
		slog.Error("Unable to delete booking", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	// Deletion ok
	slog.Info("Successfully deleted booking", "BookingId", p.Id)
	response.SendResponse(c, a, response.NewOkResponse(message.Header.MessageId))
}

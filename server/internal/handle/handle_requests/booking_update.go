package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func BookingUpdate(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get message payload unmarshalled
	var p request.BookingModifyPayload
	if err := p.UnmarshalBinary(message.Payload); err != nil {
		slog.Error("Unable to unmarshall BookingModifyPayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	// Get manager
	m := bookings.GetManager()
	err := m.UpdateBookingFromId(p.Id, p.DeltaHour)
	if err != nil {
		slog.Error("Unable to update booking", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
	}

	// Booking has been updated
	slog.Info("Booking has been updated", "BookingId", p.Id)
	response.SendResponse(c, a, response.NewOkResponse(message.Header.MessageId))
}

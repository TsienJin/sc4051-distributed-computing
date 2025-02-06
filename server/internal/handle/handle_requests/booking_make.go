package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func BookingMake(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get message payload unmarshalled
	var p request.BookingMakePayload
	if err := p.UnmarshalBinary(message.Payload); err != nil {
		slog.Error("Unable to unmarshall BookingMakePayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	booking, err := p.GetBooking()
	if err != nil {
		slog.Error("Unable to create instance of booking", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	// Get manager
	m := bookings.GetManager()
	if err := m.NewBooking(p.Name, booking); err != nil {
		slog.Error("Unable to create new Booking", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	slog.Info("Successfully made booking", "Booking", p)
	response.SendResponse(c, a, response.NewOkResponse(message.Header.MessageId))
}

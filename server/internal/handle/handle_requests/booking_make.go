package handle_requests

import (
	"fmt"
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
	if err := p.UnmarshalBinary(message.Payload[1:]); err != nil {
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

	manager := bookings.GetManager()
	if err := manager.NewBooking(p.Name, booking); err != nil {
		slog.Error("Unable to make booking", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusBadRequest, err.Error()))
		return
	}

	slog.Info("Booking has been made", "BookingId", fmt.Sprintf("%v", booking.Id))

	res := response.NewResponse(
		response.WithStatusCode(response.StatusOk),
		response.WithOriginalMessageId(message.Header.MessageId),
		response.WithPayloadBytes(booking.GetIdAsBytes()),
	)

	slog.Info("Successfully made booking", "Booking", p)
	response.SendResponse(c, a, res)
}

package handle_requests

import (
	"fmt"
	"log/slog"
	"net"
	"server/internal/bookings"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
	"time"
)

func FacilityMonitor(c *net.UDPConn, a *net.UDPAddr, message *protocol.Message) {

	// Get request payload
	var p request.FacilityMonitorPayload
	if err := p.UnmarshalBinary(message.Payload[1:]); err != nil {
		slog.Error("Unable to unmarshall FacilityMonitorPayload", "err", err)
		response.SendResponse(c, a, response.NewErrorResponse(message.Header.MessageId, response.StatusInternalServerError, err.Error()))
		return
	}

	// Register connection as a client
	go func() {
		response.SendResponse(c, a, response.NewResponse(
			response.WithOriginalMessageId(message.Header.MessageId),
			response.WithStatusCode(response.StatusOk),
			response.WithPayloadMessage(fmt.Sprintf("Monitoring %s for %d seconds", p.Name, p.Ttl)),
		))
		consumer := bookings.GetMonitor().Watch(p.Name, time.Duration(p.Ttl)*time.Second)

		// Continuously listen for messages and send them to client
		for {
			select {
			case s, ok := <-consumer.Channel:
				if !ok {
					// Channel closed, exit gracefully
					response.SendResponse(c, a, response.NewResponse(
						response.WithOriginalMessageId(message.Header.MessageId),
						response.WithStatusCode(response.StatusOk),
						response.WithPayloadMessage("Monitoring stopped (channel closed)"),
					))
					return
				}
				response.SendResponse(c, a, response.NewResponse(
					response.WithOriginalMessageId(message.Header.MessageId),
					response.WithStatusCode(response.StatusOk),
					response.WithPayloadMessage(s),
				))
			case <-consumer.Ctx.Done():
				response.SendResponse(c, a, response.NewResponse(
					response.WithOriginalMessageId(message.Header.MessageId),
					response.WithStatusCode(response.StatusOk),
					response.WithPayloadMessage("Monitoring over"),
				))
				return
			}
		}

	}()

}

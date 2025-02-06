package handle_requests

import (
	"log/slog"
	"net"
	"server/internal/protocol"
	"server/internal/rpc/request"
	"server/internal/rpc/response"
)

func Sort(c *net.UDPConn, a *net.UDPAddr, m *protocol.Message) {

	h := response.GetResponseHistoryInstance()

	// Check if message has been processed or is processing
	if done, exists := h.Check(m.Header.MessageId); exists {
		if !done {
			slog.Info("Request is supposed to invoke a processes that is still running, ignoring duplicate")
			return
		}

		slog.Info("Request received, but request has already been sent, resending response")
		r, err := h.GetResponse(m.Header.MessageId)
		if err != nil {
			slog.Error("Unable to retrieve historical response", "err", err)
			return
		}
		response.SendResponse(c, a, r)
		return
	}

	// Set processing here as only requests will have message responses
	h.SetProcessing(m.Header.MessageId)

	var req request.Request
	if err := req.UnmarshalBinary(m.Payload); err != nil {
		slog.Error("Unable to determine target method from message", "MessageId", m.Header.MessageId)
		return
	}

	switch req.MethodIdentifier {
	case request.MethodIdentifierFacilityCreate:
		FacilityCreate(c, a, m)
		break
	case request.MethodIdentifierFacilityQuery:
		FacilityQuery(c, a, m)
		break
	case request.MethodIdentifierFacilityMonitor:
		FacilityMonitor(c, a, m)
		break
	case request.MethodIdentifierFacilityDelete:
		FacilityMonitor(c, a, m)
		break
	case request.MethodIdentifierBookingMake:
		BookingMake(c, a, m)
		break
	case request.MethodIdentifierBookingUpdate:
		BookingUpdate(c, a, m)
		break
	case request.MethodIdentifierBookingDelete:
		BookingDelete(c, a, m)
		break
	default:
		slog.Error("Request type not supported", "RequestType", req.MethodIdentifier)
		return
	}

}

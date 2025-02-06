package response

import "server/internal/protocol/proto_defs"

func NewErrorResponse(mid proto_defs.MessageId, code StatusCode, message string) *Response {
	return NewResponse(
		WithOriginalMessageId(mid),
		WithStatusCode(code),
		WithPayloadBytes([]byte(message)),
	)
}

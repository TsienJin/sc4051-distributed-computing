package response

import "server/internal/protocol/proto_defs"

func NewOkResponse(mid proto_defs.MessageId) *Response {
	return NewResponse(
		WithOriginalMessageId(mid),
		WithStatusCode(StatusOk),
		WithPayloadBytes([]byte{}),
	)
}

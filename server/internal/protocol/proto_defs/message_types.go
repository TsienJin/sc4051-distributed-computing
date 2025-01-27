package proto_defs

type MessageType uint8

const (
	MessageTypeError MessageType = 1 + iota
	MessageTypeRequest
	MessageTypeResponse
	MessageTypeAcknowledge
	MessageTypeRequestResend
)

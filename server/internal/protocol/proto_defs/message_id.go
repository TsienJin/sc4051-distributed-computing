package proto_defs

import "github.com/google/uuid"

type MessageId [16]byte

func NewMessageId() MessageId {
	var id [16]byte
	u, _ := uuid.NewUUID()
	copy(id[:], u[:])
	return id
}

package proto_defs

import "time"

// PacketSizeLimit is the hard limit for header + payload for this protocol
const PacketSizeLimit int = 2 << 9

// PacketHeaderSize is the number of bytes used by the header
const PacketHeaderSize = 24

// PacketPayloadSizeLimit is the maximum allowable size in bytes for the payload
const PacketPayloadSizeLimit = PacketSizeLimit - PacketHeaderSize

// MessageTimeout is the absolute maximum time allowed to wait for an ack
const MessageTimeout time.Duration = time.Millisecond * 100

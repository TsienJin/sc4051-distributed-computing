package proto_defs

// PacketSizeLimit is the hard limit for header + payload for this protocol
const PacketSizeLimit int = 2 << 9

// PacketHeaderSize is the number of bytes used by the header
const PacketHeaderSize = 24

// PacketPayloadSizeLimit is the maximum allowable size in bytes for the payload
const PacketPayloadSizeLimit = PacketSizeLimit - PacketHeaderSize

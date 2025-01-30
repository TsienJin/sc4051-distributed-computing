package protocol

import (
	"encoding/binary"
	"hash/crc32"
)

func MakeChecksum(data []byte) uint32 {
	return crc32.Checksum(data, crc32.IEEETable)
}

func GetChecksumFromChecksumBytes(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func ValidateChecksum(data []byte, checksum uint32) bool {
	return crc32.Checksum(data, crc32.IEEETable) == checksum
}

func ValidateChecksumBytes(data []byte, checksum []byte) bool {
	return ValidateChecksum(data, binary.BigEndian.Uint32(checksum))
}

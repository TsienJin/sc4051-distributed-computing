package protocol

import "hash/crc32"

func MakeChecksum(data []byte) uint32 {
	return crc32.Checksum(data, crc32.IEEETable)
}

package request

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	MethodIdentifier MethodIdentifier
	Payload          []byte
}

func (r *Request) UnmarshalBinary(data []byte) error {
	r.MethodIdentifier = MethodIdentifier(data[0])
	r.Payload = data[1:]
	return nil
}

func (r *Request) MarshalBinary() ([]byte, error) {

	buf := &bytes.Buffer{}

	if err := binary.Write(buf, binary.BigEndian, r.MethodIdentifier); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, r.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

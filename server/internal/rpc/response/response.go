package response

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"log/slog"
	"server/internal/protocol/proto_defs"
)

type Response struct {
	OriginalMessageId proto_defs.MessageId
	StatusCode        StatusCode
	Payload           []byte
}

type Option func(*Response)

func NewResponse(opts ...Option) *Response {
	r := &Response{}
	for _, o := range opts {
		o(r)
	}
	return r
}

func WithOriginalMessageId(id proto_defs.MessageId) Option {
	return func(r *Response) {
		r.OriginalMessageId = id
	}
}

func WithStatusCode(s StatusCode) Option {
	return func(r *Response) {
		r.StatusCode = s
	}
}

func WithPayload(p encoding.BinaryMarshaler) Option {
	return func(r *Response) {
		data, err := p.MarshalBinary()
		if err != nil {
			slog.Error("Unable to marshal payload into binary", "Payload", p)
			return
		}
		r.Payload = data
	}
}

func (r *Response) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, r.OriginalMessageId); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, r.StatusCode); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, r.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

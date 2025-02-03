package request

type Request struct {
	MethodIdentifier MethodIdentifier
	Payload          []byte
}

func (r *Request) UnmarshalBinary(data []byte) error {
	r.MethodIdentifier = MethodIdentifier(data[0])
	r.Payload = data[1:]
	return nil
}

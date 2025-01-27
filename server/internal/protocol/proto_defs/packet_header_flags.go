package proto_defs

type Flags uint8

// Flags are defined in the order of LSB
const (
	FlagAckRequired Flags = 1 << iota
	FlagFragment
)

func NewFlags(flags ...Flags) Flags {
	var r Flags
	for _, f := range flags {
		r = r | f
	}
	return r
}

func (f *Flags) AckRequired() bool {
	return *f&FlagAckRequired == 1
}

func (f *Flags) Fragment() bool {
	return *f&FlagFragment == 1
}

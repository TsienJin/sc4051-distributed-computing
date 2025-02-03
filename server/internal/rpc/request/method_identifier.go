package request

type MethodIdentifier uint8

const (
	FacilityCreate  MethodIdentifier = 0x01
	FacilityQuery   MethodIdentifier = 0x02
	FacilityMonitor MethodIdentifier = 0x03
	FacilityDelete  MethodIdentifier = 0x04

	BookingMake   MethodIdentifier = 0x11
	BookingUpdate MethodIdentifier = 0x12
	BookingDelete MethodIdentifier = 0x13
)

package request

type MethodIdentifier uint8

const (
	MethodIdentifierFacilityCreate  MethodIdentifier = 0x01
	MethodIdentifierFacilityQuery   MethodIdentifier = 0x02
	MethodIdentifierFacilityMonitor MethodIdentifier = 0x03
	MethodIdentifierFacilityDelete  MethodIdentifier = 0x04

	MethodIdentifierBookingMake   MethodIdentifier = 0x11
	MethodIdentifierBookingUpdate MethodIdentifier = 0x12
	MethodIdentifierBookingDelete MethodIdentifier = 0x13
)

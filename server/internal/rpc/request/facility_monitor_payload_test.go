package request

import (
	"github.com/google/go-cmp/cmp"
	"server/internal/bookings"
	"testing"
)

func TestFacilityMonitorPayload_MarshalUnmarshalBinary(t *testing.T) {

	payload := FacilityMonitorPayload{
		Name: bookings.FacilityName("TestFacilityMonitorPayload_MarshalUnmarshalBinary"),
		Ttl:  50,
	}

	payloadBytes, err := payload.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	var reflect FacilityMonitorPayload
	if err := reflect.UnmarshalBinary(payloadBytes); err != nil {
		t.Error(err)
	}

	if !cmp.Equal(reflect, payload) {
		t.Logf("E: %v", payload)
		t.Logf("R: %v", reflect)
		t.Error("reflected FacilityMonitorPayload does not match original")
	}

}

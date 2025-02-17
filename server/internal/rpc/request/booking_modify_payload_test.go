package request

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestBookingModifyPayload_MarshalUnmarshalBinary(t *testing.T) {

	payload := &BookingModifyPayload{
		Id:        2532,
		DeltaHour: 925023,
	}

	bin, err := payload.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	var reflected BookingModifyPayload
	if err := reflected.UnmarshalBinary(bin); err != nil {
		t.Error(err)
	}

	if !cmp.Equal(reflected, *payload) {
		t.Logf("E: %v", *payload)
		t.Logf("R: %v", reflected)
		t.Error("Reflected payload does not match original")
	}

}

func TestBookingModifyPayload_MarshalUnmarshalBinary_NegativeTime(t *testing.T) {

	payload := &BookingModifyPayload{
		Id:        2532,
		DeltaHour: -925023,
	}

	bin, err := payload.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	var reflected BookingModifyPayload
	if err := reflected.UnmarshalBinary(bin); err != nil {
		t.Error(err)
	}

	if !cmp.Equal(reflected, *payload) {
		t.Logf("E: %v", *payload)
		t.Logf("R: %v", reflected)
		t.Error("Reflected payload does not match original")
	}

}

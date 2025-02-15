package integration_suite

import (
	"server/internal/rpc/request/request_constructor"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"testing"
	"time"
)

func TestCreateBooking_successful(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestCreateFacility_successful"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(5)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	if err := c.SendRpcRequestConstructors(
		request_constructor.NewFacilityCreatePacket("TestCreateBooking_successful"),
		request_constructor.NewBookingMakePacket("TestCreateBooking_successful", time.Now(), time.Now().Add(time.Duration(3)*time.Hour)),
	); err != nil {
		t.Error(err)
	}

	firstPacketOk := false
	secondPacketOk := false
	resCount := 0

LOOP:
	for {
		select {
		case <-c.Ctx.Done():
			break LOOP
		case r := <-c.Responses: // we should only expect 2 response (ok then error)

			switch resCount {
			case 0: // First response
				if r.StatusCode == response.StatusOk {
					resCount++
					firstPacketOk = true
				} else {
					t.Error("Expected first packet to be ok")
				}
			case 1: // Second packet
				if r.StatusCode == response.StatusOk {
					resCount++
					secondPacketOk = true
					break LOOP
				} else {
					t.Error("Expected second packet to be ok")
				}
			default:
				t.Error("ResCount out of range")
			}
		default:
			continue
		}
	}

	if !firstPacketOk || !secondPacketOk {
		t.Error("Test did not pass, check logs")
	}

}

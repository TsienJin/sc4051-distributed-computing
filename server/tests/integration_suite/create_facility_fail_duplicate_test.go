package integration_suite

import (
	"server/internal/rpc/request"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"testing"
	"time"
)

func TestCreateFacility_fail_duplicate(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestCreateFacility_successful"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(5)*time.Second),
	)
	defer c.Close()

	if err := c.SendRpcRequestConstructors(
		request.NewFacilityCreatePacket("TestCreateFacility_fail_duplicate"),
		request.NewFacilityCreatePacket("TestCreateFacility_fail_duplicate"),
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
				if r.StatusCode == response.StatusBadRequest {
					resCount++
					secondPacketOk = true
					break LOOP
				} else {
					t.Error("Expected second packet to be error")
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

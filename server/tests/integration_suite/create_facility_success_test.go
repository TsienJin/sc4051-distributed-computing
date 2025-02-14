package integration_suite

import (
	"server/internal/rpc/request"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"testing"
	"time"
)

func TestCreateFacility_successful(t *testing.T) {

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
		request.NewFacilityCreatePacket("TestCreateFacility_successful"),
	); err != nil {
		t.Error(err)
	}

	ok := false

LOOP:
	for {
		select {
		case <-c.Ctx.Done():
			break LOOP
		case r := <-c.Responses: // we should only expect 1 response (ok or error)
			ok = r.StatusCode == response.StatusOk
			break LOOP
		default:
			continue
		}
	}

	if !ok {
		t.Error("Test did not pass, check if response was received")
	}

}

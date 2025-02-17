package integration_suite

import (
	"server/internal/rpc/request/request_constructor"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"server/tests/test_response"
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
		client.WithTimeout(time.Duration(15)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	if err := c.SendRpcRequestConstructors(
		request_constructor.NewFacilityCreatePacket("TestCreateFacility_successful"),
	); err != nil {
		t.Error(err)
	}

	c.ValidateResponses(t, test_response.BeStatus(response.StatusOk))

}

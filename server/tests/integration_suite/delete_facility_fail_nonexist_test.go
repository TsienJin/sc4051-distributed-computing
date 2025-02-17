package integration_suite

import (
	"server/internal/client"
	"server/internal/interfaces"
	"server/internal/rpc/request/request_constructor"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/test_response"
	"testing"
	"time"
)

func TestDeleteFacility_fail_non_exists(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestDeleteFacility_fail_non_exists"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(15)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	c.SendSyncWithValidator(
		t,
		[]interfaces.RpcRequestConstructor{
			request_constructor.NewFacilityDeletePacket("TestDeleteFacility_fail_non_exists"),
		},
		[]test_response.ResponseValidator{
			test_response.BeStatus(response.StatusBadRequest),
		},
	)

}

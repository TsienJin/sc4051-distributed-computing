package integration_suite

import (
	"server/internal/interfaces"
	"server/internal/rpc/request/request_constructor"
	"server/internal/rpc/response"
	"server/internal/server"
	"server/tests/client"
	"server/tests/test_response"
	"testing"
	"time"
)

func TestCreateBooking_fail_duplicate(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestCreateBooking_fail_duplicate"),
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
			request_constructor.NewFacilityCreatePacket("TestCreateBooking_fail_duplicate"),
			request_constructor.NewBookingMakePacket("TestCreateBooking_fail_duplicate", time.Now(), time.Now().Add(time.Duration(3)*time.Hour)),
			request_constructor.NewBookingMakePacket("TestCreateBooking_fail_duplicate", time.Now(), time.Now().Add(time.Duration(3)*time.Hour))},
		[]test_response.ResponseValidator{
			test_response.BeStatus(response.StatusOk),
			test_response.BeStatus(response.StatusOk),
			test_response.BeStatus(response.StatusBadRequest)},
	)

}

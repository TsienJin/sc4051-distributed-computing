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

func TestDeleteBooking_successful(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestDeleteBooking_successful"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(15)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	bidChan := make(chan uint16, 1)
	deleteSuccessful := false

	defer func() {
		if !deleteSuccessful {
			t.Error("Client did not manage to delete booking")
		}
	}()

	c.SendSyncWithValidator(
		t,
		[]interfaces.RpcRequestConstructor{
			request_constructor.NewFacilityCreatePacket("TestDeleteBooking_successful"),
			request_constructor.NewBookingMakePacket("TestDeleteBooking_successful", time.Now(), time.Now().Add(time.Duration(3)*time.Hour))},
		[]test_response.ResponseValidator{
			test_response.BeStatus(response.StatusOk),
			test_response.PacketMustPassAll(
				test_response.BeStatus(response.StatusOk),
				test_response.ExtractBookingId(bidChan),
			),
		},
	)

	c.SendSyncWithValidator(
		t,
		[]interfaces.RpcRequestConstructor{request_constructor.NewBookingDeletePacket(<-bidChan)},
		[]test_response.ResponseValidator{test_response.BeStatus(response.StatusOk)},
	)

	deleteSuccessful = true

}

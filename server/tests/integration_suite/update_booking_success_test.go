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

func TestUpdateFacility_successful_positive_delta(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestUpdateFacility_successful_positive_delta"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(15)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	bidChan := make(chan uint16, 1)
	modifySuccessful := false

	defer func() {
		if !modifySuccessful {
			t.Error("Client did not manage to update booking")
		}
	}()

	c.SendSyncWithValidator(
		t,
		[]interfaces.RpcRequestConstructor{
			request_constructor.NewFacilityCreatePacket("TestUpdateFacility_successful_positive_delta"),
			request_constructor.NewBookingMakePacket("TestUpdateFacility_successful_positive_delta", time.Now(), time.Now().Add(time.Duration(3)*time.Hour))},
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
		[]interfaces.RpcRequestConstructor{request_constructor.NewBookingModifyPacket(<-bidChan, 5)},
		[]test_response.ResponseValidator{test_response.BeStatus(response.StatusOk)},
	)

	modifySuccessful = true
}

func TestUpdateFacility_successful_negative_delta(t *testing.T) {

	serverPort, err := server.ServeRandomPort()
	if err != nil {
		t.Error(err)
	}

	c, err := client.NewClient(
		client.WithClientName("TestUpdateFacility_successful_negative_delta"),
		client.WithTargetAsIpV4("127.0.0.1", serverPort),
		client.WithTimeout(time.Duration(5)*time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	bidChan := make(chan uint16, 1)
	modifySuccessful := false

	defer func() {
		if !modifySuccessful {
			t.Error("Client did not manage to update booking")
		}
	}()

	c.SendSyncWithValidator(
		t,
		[]interfaces.RpcRequestConstructor{
			request_constructor.NewFacilityCreatePacket("TestUpdateFacility_successful_negative_delta"),
			request_constructor.NewBookingMakePacket("TestUpdateFacility_successful_negative_delta", time.Now(), time.Now().Add(time.Duration(3)*time.Hour))},
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
		[]interfaces.RpcRequestConstructor{request_constructor.NewBookingModifyPacket(<-bidChan, -5)},
		[]test_response.ResponseValidator{test_response.BeStatus(response.StatusOk)},
	)

	modifySuccessful = true
}

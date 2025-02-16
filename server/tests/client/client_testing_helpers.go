package client

import (
	"server/internal/interfaces"
	"server/tests/test_response"
	"testing"
)

// ValidateResponses is a function to validate each packet that comes in corresponding to the index of the response
// validator.
func (c *Client) ValidateResponses(t *testing.T, rv ...test_response.ResponseValidator) {

	if len(rv) == 0 {
		t.Error("No validators passed to ValidateResponse")
	}

	packetCount := 0

LOOP:
	for {
		select {
		case <-c.Ctx.Done():
			if packetCount != len(rv) {
				t.Errorf("Expected %d packets, but only validated %d", len(rv), packetCount)
			}
			break LOOP
		case r := <-c.Responses:

			if packetCount >= len(rv) {
				t.Error("Unexpected packet to validate")
			}

			if err := rv[packetCount](r); err != nil {
				t.Error(err)
			}
			packetCount++

			if packetCount == len(rv) {
				break LOOP
			}
		default:
			continue
		}
	}
}

func (c *Client) SendSyncWithValidator(
	t *testing.T,
	constructors []interfaces.RpcRequestConstructor,
	validators []test_response.ResponseValidator,
) {

	// Sanity check to ensure that tests match validators
	if len(constructors) != len(validators) {
		t.Error("Constructors and validators do not match in length")
	}

	number := len(constructors)
	for i := 0; i < number; i++ {
		if err := c.SendRpcRequestConstructors(constructors[i]); err != nil {
			t.Error(err)
			break
		}
		c.ValidateResponses(t, validators[i])
	}

}

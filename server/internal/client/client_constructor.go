package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/response"
	"server/tests"
	"time"
)

type NewClientOpt func(*Client)

func WithTarget(t *net.UDPAddr) NewClientOpt {
	return func(c *Client) {
		c.targetServer = t
	}
}

func WithTargetAsIpV4(host string, port int) NewClientOpt {
	return func(c *Client) {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return
		}
		c.targetServer = addr
	}
}

func WithClientName(name string) NewClientOpt {
	return func(c *Client) {
		if name == "" {
			c.name = "CLIENT"
			return
		}
		c.name = name
	}
}

func WithTimeout(t time.Duration) NewClientOpt {
	return func(c *Client) {
		c.Ctx, c.Cancel = context.WithTimeout(c.Ctx, t)
	}
}

func WithCustomLogger(logger *slog.Logger) NewClientOpt {
	return func(c *Client) {
		c.logger = logger
	}
}

func NewClient(opts ...NewClientOpt) (*Client, error) {
	outChan := make(chan *response.Response, 8)

	c := &Client{
		sequencer:     newResponseSequencer(outChan),
		responseBytes: make(chan [proto_defs.PacketSizeLimit]byte, 8),
		Responses:     outChan,
	}
	c.Ctx, c.Cancel = context.WithCancel(context.Background())
	c.logger = tests.NewNamedTestLogger(c.name)
	for _, o := range opts {
		o(c)
	}

	c.manager = newSendHistory()
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := c.createConn(); err != nil {
		return nil, err
	}

	c.wg.Add(2)
	go c.receivePacketLoop()
	go c.handleIncomingPacket()

	return c, nil
}

func (c *Client) validate() error {
	if c.targetServer == nil {
		return errors.New("client target missing")
	}

	if !reflect.ValueOf(c.responseBytes).IsValid() || reflect.ValueOf(c.responseBytes).IsNil() {
		return errors.New("client responses chan is missing")
	}

	return nil
}

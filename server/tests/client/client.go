package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/tests"
	"sync"
	"time"
)

type Client struct {
	wg sync.WaitGroup

	logger *slog.Logger
	name   string

	history      *sendHistory
	targetServer *net.UDPAddr
	conn         *net.UDPConn

	responseBytes chan [proto_defs.PacketSizeLimit]byte // This chan is used internally for message passing
	Responses     chan *protocol.Packet                 // Exposed to process incoming messages

	Ctx    context.Context
	Cancel context.CancelFunc
}

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

func NewClient(opts ...NewClientOpt) (*Client, error) {
	c := &Client{
		responseBytes: make(chan [proto_defs.PacketSizeLimit]byte, 8),
		Responses:     make(chan *protocol.Packet, 8),
	}
	c.Ctx, c.Cancel = context.WithCancel(context.Background())
	c.logger = tests.NewNamedTestLogger(c.name)
	for _, o := range opts {
		o(c)
	}

	c.history = newSendHistory()
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

func (c *Client) createConn() error {
	// Resolve a UDP address with port 0 (OS assigns a free port)
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Listen on a random available port
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on a UDP port: %w", err)
	}

	c.conn = conn

	return nil
}

func (c *Client) receivePacketLoop() {
	defer c.wg.Done()

	buffer := make([]byte, proto_defs.PacketSizeLimit)
	for {
		select {
		case <-c.Ctx.Done():
			c.logger.Info("Context closed, exiting 'receivePacketLoop'")
			return
		default:
			n, _, err := c.conn.ReadFromUDP(buffer)
			if err != nil {
				c.logger.Error("Unable to read bytes from UDP connection")
				continue
			}
			res := [proto_defs.PacketSizeLimit]byte{}
			copy(res[:n], buffer[:n])

			// Send bytes to chan
			c.responseBytes <- res
		}
	}
}

func (c *Client) handleIncomingPacket() {
	defer c.wg.Done()

	for {
		select {
		case <-c.Ctx.Done():
			c.logger.Info("Context closed, exiting 'handleIncomingPacket")
			return
		case data := <-c.responseBytes:
			// Unmarshal binary and validate
			var p protocol.Packet
			err := p.UnmarshalBinary(data[:])
			if err != nil {
				c.logger.Error("Unable to unmarshal binary", "err", err)
				return
			}

			switch p.Header.MessageType {

			case proto_defs.MessageTypeAcknowledge:
				var ackPayload protocol.AckResendPayload
				if err := ackPayload.UnmarshalBinary(p.Payload); err != nil {
					c.logger.Error("Unable to unmarshal Ack payload", "err", err)
					continue
				}
				// Remove from history
				ident := ackPayload.ToPacketIdent()
				c.logger.Debug("Ack received for packet, removing from history", "ident", ident)
				c.history.clear(ident)

			case proto_defs.MessageTypeRequestResend:
				var resendPayload protocol.AckResendPayload
				if err := resendPayload.UnmarshalBinary(p.Payload); err != nil {
					c.logger.Error("Unable to unmarshal Resend payload", "err", err)
					continue
				}
				// Resend from history
				ident := resendPayload.ToPacketIdent()
				c.logger.Debug("Resend request received", "ReqIdent", ident)
				packet, err := c.history.get(ident)
				if err != nil {
					c.logger.Error("Unable to retrieve packet from history", "err", err)
					continue
				}
				if err := c.SendPacket(packet); err != nil {
					c.logger.Error("Unable to resend packet", "err", err)
					continue
				}
			case proto_defs.MessageTypeResponse:
				c.Responses <- &p
				continue
			default: // Unrecognised message types + requests
				c.logger.Warn("Unsupported packet type", "type", p.Header.MessageType)
			}

		default:
			continue
		}
	}

}

func (c *Client) Close() {
	c.logger.Debug("Closing client")
	c.Cancel()
	c.conn.Close()
	c.wg.Wait()
}

func (c *Client) SendPacket(p *protocol.Packet) error {

	c.logger.Info("Sending packet", "packet", p)

	b, err := p.MarshalBinary()
	if err != nil {
		c.logger.Error("Unable to marshal packet binary", "err", err)
		return err
	}

	if _, err := c.conn.WriteToUDP(b, c.targetServer); err != nil {
		c.logger.Error("Unable to send packet to target", "err", err)
		return err
	}

	c.history.set(p)

	return nil
}

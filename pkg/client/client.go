package client

import (
	"net"
	"time"

	"github.com/alessio-palumbo/hikari/internal/protocol"
)

const (
	lifxPort       = 56700
	recvBufferSize = 1024

	defaultSource   uint32 = 0x00000000
	defaultDeadline        = 2 * time.Second
)

type Client struct {
	conn     *net.UDPConn
	source   uint32
	deadline time.Duration
}

type Config struct {
	Source   uint32
	Deadline time.Duration
}

// HandlerFunc processes a received message and address.
type HandlerFunc func(*protocol.Message, *net.UDPAddr)

// NewClient creates a new LIFX client with the specified configuration.
// If cfg is nil, default values will be used for source and deadline.
func NewClient(cfg *Config) (*Client, error) {
	addr := &net.UDPAddr{Port: 0, IP: net.IPv4zero}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	source := defaultSource
	deadline := defaultDeadline
	if cfg != nil {
		if cfg.Source != 0 {
			source = cfg.Source
		}
		if cfg.Deadline > 0 {
			deadline = cfg.Deadline
		}
	}

	return &Client{
		conn:     conn,
		source:   source,
		deadline: deadline,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Send sends a message to the specified destination address.
func (c *Client) Send(dst *net.UDPAddr, msg *protocol.Message) error {
	msg.SetSource(c.source)

	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(data, dst)
	return err
}

// SendBroadcast sends a message to the broadcast address for LIFX devices.
func (c *Client) SendBroadcast(msg *protocol.Message) error {
	msg.SetTarget(protocol.TargetBroadcast)
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: lifxPort,
	}
	return c.Send(broadcastAddr, msg)
}

// Receive reads UDP messages until the deadline is hit or recvOne is true and a valid message has been received.
func (c *Client) Receive(timeout time.Duration, recvOne bool, handler HandlerFunc) error {
	if timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
		// Reset deadline after reading
		defer c.conn.SetReadDeadline(time.Time{})
	}

	buf := make([]byte, recvBufferSize)

	for {
		n, addr, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return err
		}

		var msg protocol.Message
		if err := msg.UnmarshalBinary(buf[:n]); err != nil {
			// skip malformed
			continue
		}

		handler(&msg, addr)
		if recvOne {
			return nil
		}
	}

	return nil
}

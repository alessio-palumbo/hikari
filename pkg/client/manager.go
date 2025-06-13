package client

import (
	"fmt"
	"net"
	"time"

	"github.com/alessio-palumbo/hikari/gen/protocol/enums"
	"github.com/alessio-palumbo/hikari/gen/protocol/packets"
	"github.com/alessio-palumbo/hikari/internal/protocol"
)

const defaultRecvBufferSize = 10

type DeviceManager struct {
	client   *Client
	sessions map[string]*DeviceSession
	recvDone chan struct{}
}

// NewDeviceManager creates a new DeviceManager that starts listening for devices.
func NewDeviceManager() (*DeviceManager, error) {
	client, err := NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	dm := &DeviceManager{
		client:   client,
		sessions: make(map[string]*DeviceSession),
		recvDone: make(chan struct{}),
	}
	go dm.recvloop()

	if err := dm.Discover(); err != nil {
		return nil, fmt.Errorf("failed to discover devices: %w", err)
	}
	return dm, nil
}

// Close closes the DeviceManager, stopping the recv loop and closing all device sessions.
func (d *DeviceManager) Close() error {
	// Close the client connection and wait for the recv loop to finish.
	d.client.conn.SetDeadline(time.Now())
	<-d.recvDone
	d.client.Close()

	for _, session := range d.sessions {
		if err := session.Close(); err != nil {
			return fmt.Errorf("failed to close device session: %w", err)
		}
	}
	clear(d.sessions)
	return nil
}

func (d *DeviceManager) Discover() error {
	msg := protocol.NewMessage(&packets.DeviceGetService{})
	return d.client.SendBroadcast(msg)
}

func (d *DeviceManager) AddSession(addr *net.UDPAddr, target [8]byte) error {
	device := NewDevice(addr, target)
	session, err := NewDeviceSession(device, d.client)
	if err != nil {
		return fmt.Errorf("failed to create device session: %w", err)
	}
	d.sessions[device.Address.IP.String()] = session

	go session.recvloop()
	go session.run()

	return nil
}

func (d *DeviceManager) PrintSessions() {
	for _, session := range d.sessions {
		fmt.Println(session.device.String())
	}
}

// recv listens for incoming messages from devices and dispatches them to the appropriate session.
func (d *DeviceManager) recvloop() {
	defer close(d.recvDone)
	buf := make([]byte, recvBufferSize)

	for {
		n, addr, err := d.client.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			panic(fmt.Sprintf("failed to read from UDP: %v", err))
		}

		var msg protocol.Message
		if err := msg.UnmarshalBinary(buf[:n]); err != nil {
			// skip malformed
			continue
		}

		if session, ok := d.sessions[addr.IP.String()]; ok {
			select {
			case session.inbound <- &msg:
			default:
				// If the channel is full, we skip the message to avoid blocking.
			}
		} else if state, ok := msg.Payload.(*packets.DeviceStateService); ok && state.Service == enums.DeviceServiceDEVICESERVICEUDP {
			if err := d.AddSession(addr, msg.Header.Target); err != nil {
				fmt.Println("Failed to spin device worker:", err)
			}
		}
	}
}

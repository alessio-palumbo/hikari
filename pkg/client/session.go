package client

import (
	"fmt"
	"net"

	"github.com/alessio-palumbo/hikari/gen/protocol/packets"
	"github.com/alessio-palumbo/hikari/internal/protocol"
)

// sender is an interface that defines the Send method for sending messages.
type sender interface {
	Send(dst *net.UDPAddr, msg *protocol.Message) error
}

// DeviceSession represents a session for a specific device.
type DeviceSession struct {
	sender  sender
	device  *Device
	inbound chan *protocol.Message // Map of sequence number to response channel
	seq     uint8                  // Sequence number for messages
	done    chan struct{}
}

// NewDeviceSession creates a new DeviceSession for the given device.
func NewDeviceSession(device *Device, sender sender) (*DeviceSession, error) {
	return &DeviceSession{
		sender:  sender,
		device:  device,
		inbound: make(chan *protocol.Message, defaultRecvBufferSize),
		done:    make(chan struct{}),
	}, nil
}

// Close closes the DeviceSession, stopping the recv loop and cleaning up resources.
func (s *DeviceSession) Close() error {
	close(s.done)
	return nil
}

// Send sends one or more messages to the device.
func (s *DeviceSession) Send(msgs ...*protocol.Message) {
	for _, msg := range msgs {
		msg.SetTarget(s.device.Serial)
		if err := s.sender.Send(s.device.Address, msg); err != nil {
			fmt.Printf("Failed to send message to device %s: %v\n", s.device.Serial, err)
		}
	}
}

// nextSeq increments the sequence number and returns the new value.
// It wraps around after reaching 255.
// nextSeq is not thread-safe and should be called with care in concurrent contexts.
func (s *DeviceSession) nextSeq() uint8 {
	s.seq++
	return s.seq
}

func (s *DeviceSession) run() error {
	s.Send(protocol.NewMessage(&packets.LightGet{}), protocol.NewMessage(&packets.DeviceGetVersion{}))
	return nil
}

// recvloop listens for incoming messages from the device and processes them.
func (s *DeviceSession) recvloop() {
	for {
		select {
		case msg := <-s.inbound:
			if msg == nil {
				continue
			}

			switch payload := msg.Payload.(type) {
			case *packets.LightState:
				if s.device.State == nil {
					s.device.State = &State{}
				}
				s.device.State.Label = string(payload.Label[:])
				s.device.State.Color = payload.Color
				s.device.State.Power = payload.Power
			case *packets.DeviceStateVersion:
				s.device.Product = payload.Product
			default:
				fmt.Println("Unhandled message type:", payload.PayloadType())
			}
		case <-s.done:
			fmt.Println("Exiting recv loop for device:", s.device.Serial)
			return
		}
	}
}

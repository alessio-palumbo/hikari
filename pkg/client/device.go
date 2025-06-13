package client

import (
	"fmt"
	"net"

	"github.com/alessio-palumbo/hikari/gen/protocol/packets"
)

type Serial [8]byte

func (s Serial) String() string {
	return fmt.Sprintf("%x", s[:6])
}

type Device struct {
	Address *net.UDPAddr
	Serial  Serial
	Product uint32
	State   *State
}

type State struct {
	Label string
	Color packets.LightHsbk
	Power uint16
}

func NewDevice(address *net.UDPAddr, serial [8]byte) *Device {
	return &Device{Address: address, Serial: Serial(serial)}
}

func (d *Device) String() string {
	return fmt.Sprintf("Device{Address: %s, Serial: %s, Product: %d, State: %s}", d.Address, d.Serial, d.Product, d.State)
}

func (d *State) String() string {
	if d == nil {
		return "State: nil"
	}
	return fmt.Sprintf("Label: %s, Color: %v, Power: %d", d.Label, d.Color, d.Power)
}

package client

import (
	"fmt"
	"math"
	"net"
	"slices"
	"strings"

	"github.com/alessio-palumbo/hikari/gen/protocol/packets"
	"github.com/alessio-palumbo/hikari/internal/protocol"
)

type deviceType string

const (
	DeviceTypeLight  deviceType = "light"
	DeviceTypeSwitch deviceType = "switch"
)

type lightType string

const (
	LightTypeSingleZone lightType = "single_zone"
	LightTypeMultiZone  lightType = "multi_zone"
	LightTypeMatrix     lightType = "matrix"
)

type Serial [8]byte

func (s Serial) String() string {
	return fmt.Sprintf("%x", s[:6])
}

func (s Serial) IsNil() bool {
	return s == [8]byte{}
}

type Device struct {
	Address         *net.UDPAddr
	Serial          Serial
	Label           string
	ProductID       uint32
	FirmwareVersion string
	Type            deviceType
	LightType       lightType
	Location        string
	Group           string
	Color           Color
	PoweredOn       bool
}

func convert16BitValueToPercentage(v uint16, multiplier float64) float64 {
	return math.Round(float64(v) * multiplier / math.MaxUint16)
}

type Color struct {
	Hue        float64
	Saturation float64
	Brightness float64
	Kelvin     uint16
}

func NewColor(hsbk packets.LightHsbk) Color {
	return Color{
		Hue:        convert16BitValueToPercentage(hsbk.Hue, 360),
		Saturation: convert16BitValueToPercentage(hsbk.Saturation, 100),
		Brightness: convert16BitValueToPercentage(hsbk.Brightness, 100),
		Kelvin:     hsbk.Kelvin,
	}
}

func (c *Color) String() string {
	if c.Saturation == 0 {
		return fmt.Sprintf("Kelvin: %d, Saturation: 0%, Brightness: %f%", c.Kelvin, c.Brightness)
	}
	return fmt.Sprintf("Hue: %f, Saturation: %f%, Brightness: %f%", c.Hue, c.Saturation, c.Brightness)
}

func NewDevice(address *net.UDPAddr, serial [8]byte) *Device {
	return &Device{Address: address, Serial: Serial(serial)}
}

func SortDevices(devices []Device) {
	slices.SortFunc(devices, func(a, b Device) int {
		if n := strings.Compare(a.Label, b.Label); n != 0 {
			return n
		}
		// If names are equal, order by serial
		return strings.Compare(a.Serial.String(), b.Serial.String())
	})
}

func DeviceStateMessages() []*protocol.Message {
	return []*protocol.Message{
		protocol.NewMessage(&packets.DeviceGetVersion{}),
		protocol.NewMessage(&packets.DeviceGetLabel{}),
		protocol.NewMessage(&packets.LightGet{}),
		protocol.NewMessage(&packets.DeviceGetHostFirmware{}),
	}
}

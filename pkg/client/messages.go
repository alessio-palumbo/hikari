package client

import (
	"fmt"
	"math"
	"time"

	"github.com/alessio-palumbo/hikari/gen/protocol/enums"
	"github.com/alessio-palumbo/hikari/gen/protocol/packets"
	"github.com/alessio-palumbo/hikari/internal/protocol"
)

const (
	defaultPeriod = time.Second
)

func SetPowerOn() *protocol.Message {
	return protocol.NewMessage(&packets.DeviceSetPower{Level: math.MaxUint16})
}

func SetPowerOff() *protocol.Message {
	return protocol.NewMessage(&packets.DeviceSetPower{Level: 0})
}

func SetColor(h, s, b *float64, k *uint16, d time.Duration) *protocol.Message {
	if d < time.Second {
		d = defaultPeriod
	}
	m := &packets.LightSetWaveformOptional{
		Color:    packets.LightHsbk{},
		Waveform: enums.LightWaveformLIGHTWAVEFORMSAW,
		Cycles:   1.0,
		Period:   uint32(d.Milliseconds()),
	}
	if h != nil {
		m.Color.Hue = convertExternalToDeviceValue(*h, 360)
		fmt.Println(m.Color.Hue)
		m.SetHue = true
	}
	if s != nil {
		m.Color.Saturation = convertExternalToDeviceValue(*s, 100)
		fmt.Println(m.Color.Saturation)
		m.SetSaturation = true
	}
	if b != nil {
		m.Color.Brightness = convertExternalToDeviceValue(*b, 100)
		fmt.Println(m.Color.Brightness)
		m.SetBrightness = true
	}
	if k != nil {
		m.Color.Kelvin = *k
		fmt.Println(m.Color.Kelvin)
		m.SetKelvin = true
	}
	return protocol.NewMessage(m)
}

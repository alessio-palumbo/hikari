package decode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeProtocol(t *testing.T) {
	yamlInput := `
enums:
  PowerLevel:
    type: uint8
    values:
      - name: off
        value: 0
      - name: on
        value: 1

fields:
  ButtonTargetRelays:
    size_bytes: 16
    fields:
      - name: "RelaysCount"
        type: "uint8"
        size_bytes: 1
      - name: "Relays"
        type: "[15]uint8"
        size_bytes: 15

packets:
  light:
    LightOn:
      pkt_type: 117
      size_bytes: 0
      fields: []
`

	spec, err := DecodeProtocol([]byte(yamlInput))
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Validate Enums
	require.Len(t, spec.Enums, 1)
	require.Equal(t, "PowerLevel", spec.Enums[0].Name)
	require.Equal(t, "uint8", spec.Enums[0].Type)
	require.Len(t, spec.Enums[0].Values, 2)
	require.Equal(t, "off", spec.Enums[0].Values[0].Name)
	require.Equal(t, 0, spec.Enums[0].Values[0].Value)

	// Validate Fields
	require.Len(t, spec.Fields, 1)
	require.Equal(t, "ButtonTargetRelays", spec.Fields[0].Name)
	require.Equal(t, 16, spec.Fields[0].SizeBytes)
	require.Len(t, spec.Fields[0].Fields, 2)
	require.Equal(t, "RelaysCount", spec.Fields[0].Fields[0].Name)
	require.Equal(t, "uint8", spec.Fields[0].Fields[0].Type)
	require.Equal(t, 1, spec.Fields[0].Fields[0].SizeBytes)
	require.Equal(t, "Relays", spec.Fields[0].Fields[1].Name)
	require.Equal(t, "[15]uint8", spec.Fields[0].Fields[1].Type)
	require.Equal(t, 15, spec.Fields[0].Fields[1].SizeBytes)

	// Validate Packets
	require.Len(t, spec.Packets, 1)
	require.Equal(t, "LightOn", spec.Packets[0].Name)
	require.Equal(t, "light", spec.Packets[0].Namespace)
	require.Equal(t, 117, spec.Packets[0].PktType)
}

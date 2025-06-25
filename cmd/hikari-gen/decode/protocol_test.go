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
      - name: "Power"
        type: "<PowerLevel>"
        size_bytes: 8

unions:
  ButtonTarget:
    size_bytes: 16
    fields:
      - type: reserved
        size_bytes: 16
      - name: "Relays"
        type: "<ButtonTargetRelays>"
        size_bytes: 16
      - name: "Group"
        type: "[16]byte"
        size_bytes: 16

packets:
  button:
    ButtonSet:
      pkt_type: 906
      size_bytes: 810
      fields:
        - name: "Index"
          type: "uint8"
          size_bytes: 1
        - name: "ButtonsCount"
          type: "uint8"
          size_bytes: 1
        - name: "Buttons"
          type: "[4]<ButtonTargetRelays>"
          size_bytes: 64
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
	require.Equal(t, "on", spec.Enums[0].Values[1].Name)
	require.Equal(t, 1, spec.Enums[0].Values[1].Value)

	// Validate Fields
	require.Len(t, spec.Fields, 1)
	require.Equal(t, "ButtonTargetRelays", spec.Fields[0].Name)
	require.Equal(t, 16, spec.Fields[0].SizeBytes)
	require.Len(t, spec.Fields[0].Fields, 3)
	require.Equal(t, "RelaysCount", spec.Fields[0].Fields[0].Name)
	require.Equal(t, "uint8", spec.Fields[0].Fields[0].Type)
	require.Equal(t, 1, spec.Fields[0].Fields[0].SizeBytes)
	require.Equal(t, "Relays", spec.Fields[0].Fields[1].Name)
	require.Equal(t, "[15]uint8", spec.Fields[0].Fields[1].Type)
	require.Equal(t, 15, spec.Fields[0].Fields[1].SizeBytes)
	require.Equal(t, "Power", spec.Fields[0].Fields[2].Name)
	require.Equal(t, "enums.PowerLevel", spec.Fields[0].Fields[2].Type)
	require.Equal(t, 8, spec.Fields[0].Fields[2].SizeBytes)

	// Validate Unions
	require.Len(t, spec.Unions, 1)
	require.Equal(t, "ButtonTarget", spec.Unions[0].Name)
	require.Equal(t, 16, spec.Unions[0].SizeBytes)
	require.Len(t, spec.Unions[0].Fields, 3)
	require.Equal(t, "", spec.Unions[0].Fields[0].Name)
	require.Equal(t, "reserved", spec.Unions[0].Fields[0].Type)
	require.Equal(t, 16, spec.Unions[0].Fields[0].SizeBytes)
	require.Equal(t, "Relays", spec.Unions[0].Fields[1].Name)
	require.Equal(t, "ButtonTargetRelays", spec.Unions[0].Fields[1].Type)
	require.Equal(t, 16, spec.Unions[0].Fields[1].SizeBytes)
	require.Equal(t, "Group", spec.Unions[0].Fields[2].Name)
	require.Equal(t, "[16]byte", spec.Unions[0].Fields[2].Type)
	require.Equal(t, 16, spec.Unions[0].Fields[2].SizeBytes)

	// Validate Packets
	require.Len(t, spec.Packets, 1)
	require.Equal(t, "button", spec.Packets[0].Namespace)
	require.Equal(t, "ButtonSet", spec.Packets[0].Name)
	require.Equal(t, 906, spec.Packets[0].PktType)
	require.Len(t, spec.Packets[0].Fields, 3)
	require.Equal(t, "Index", spec.Packets[0].Fields[0].Name)
	require.Equal(t, "uint8", spec.Packets[0].Fields[0].Type)
	require.Equal(t, 1, spec.Packets[0].Fields[0].SizeBytes)
	require.Equal(t, "ButtonsCount", spec.Packets[0].Fields[1].Name)
	require.Equal(t, "uint8", spec.Packets[0].Fields[1].Type)
	require.Equal(t, 1, spec.Packets[0].Fields[1].SizeBytes)
	require.Equal(t, "Buttons", spec.Packets[0].Fields[2].Name)
	require.Equal(t, "[4]ButtonTargetRelays", spec.Packets[0].Fields[2].Type)
	require.Equal(t, 64, spec.Packets[0].Fields[2].SizeBytes)
}

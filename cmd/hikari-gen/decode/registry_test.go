package decode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeProducts(t *testing.T) {
	jsonInput := `
[
  {
    "vid": 1,
    "name": "LIFX",
    "defaults": {
      "hev": false,
      "color": false,
      "chain": false,
      "matrix": false,
      "relays": false,
      "buttons": false,
      "infrared": false,
      "multizone": false,
      "temperature_range": null,
      "extended_multizone": false
    },
    "products": [
      {
        "pid": 1,
        "name": "LIFX Original 1000",
        "features": {
          "color": true,
          "chain": false,
          "matrix": false,
          "infrared": false,
          "multizone": false,
          "temperature_range": [2500, 9000]
        },
        "upgrades": [
          {
            "major": 2,
            "minor": 80,
            "features": {
              "temperature_range": [1500, 9000]
            }
          }
        ]
      }
    ]
  }
]
`

	ps, err := DecodeProductsRegistry([]byte(jsonInput))
	require.NoError(t, err)
	require.NotNil(t, ps)

	// Validate Enums
	require.Len(t, ps, 1)
	require.Equal(t, 1, ps[0].PID)
	require.Equal(t, "LIFX Original 1000", ps[0].Name)
	require.Equal(t, true, ps[0].Features.Color)
	require.Equal(t, false, ps[0].Features.Chain)
	require.Equal(t, false, ps[0].Features.Matrix)
	require.Equal(t, false, ps[0].Features.Infrared)
	require.Equal(t, false, ps[0].Features.Multizone)
	require.Len(t, ps[0].Features.TemperatureRange, 2)
	require.Equal(t, 2500, ps[0].Features.TemperatureRange[0])
	require.Equal(t, 9000, ps[0].Features.TemperatureRange[1])
	require.Len(t, ps[0].Upgrades, 1)
	require.Equal(t, 2, ps[0].Upgrades[0].Major)
	require.Equal(t, 80, ps[0].Upgrades[0].Minor)
	require.Len(t, ps[0].Upgrades[0].Features.TemperatureRange, 2)
	require.Equal(t, 1500, ps[0].Upgrades[0].Features.TemperatureRange[0])
	require.Equal(t, 9000, ps[0].Upgrades[0].Features.TemperatureRange[1])
}

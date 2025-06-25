package decode

import (
	"encoding/json"
	"fmt"
)

type FeatureSet struct {
	HEV               bool  `json:"hev"`
	Color             bool  `json:"color"`
	Chain             bool  `json:"chain"`
	Matrix            bool  `json:"matrix"`
	Relays            bool  `json:"relays"`
	Buttons           bool  `json:"buttons"`
	Infrared          bool  `json:"infrared"`
	Multizone         bool  `json:"multizone"`
	TemperatureRange  []int `json:"temperature_range"`
	ExtendedMultizone bool  `json:"extended_multizone"`
}

type Product struct {
	PID      int        `json:"pid"`
	Name     string     `json:"name"`
	Features FeatureSet `json:"features"`
	Upgrades []Upgrade  `json:"upgrades"`
}

type Upgrade struct {
	Major    int        `json:"major"`
	Minor    int        `json:"minor"`
	Features FeatureSet `json:"features"`
}

type Vendor struct {
	VID      int        `json:"vid"`
	Name     string     `json:"name"`
	Defaults FeatureSet `json:"defaults"`
	Products []Product  `json:"products"`
}

func DecodeProductsRegistry(b []byte) ([]Product, error) {
	var vendors []Vendor
	if err := json.Unmarshal(b, &vendors); err != nil {
		return nil, fmt.Errorf("failed to parse products JSON: %w", err)
	}

	var lifxVendor *Vendor
	for _, v := range vendors {
		if v.VID == 1 {
			lifxVendor = &v
			break
		}
	}

	if lifxVendor == nil {
		return nil, fmt.Errorf("could not parse lifx vendor")
	}

	return lifxVendor.Products, nil
}

package generate

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
)

func TestGenerateProductsRegistry(t *testing.T) {
	products := []decode.Product{
		{
			PID:  1,
			Name: "LIFX Original 1000",
			Features: decode.FeatureSet{
				Color:            true,
				Chain:            false,
				Matrix:           false,
				Infrared:         false,
				Multizone:        false,
				TemperatureRange: []int{2500, 9000},
			},
			Upgrades: []decode.Upgrade{
				{
					Major: 2,
					Minor: 80,
					Features: decode.FeatureSet{
						TemperatureRange: []int{1500, 9000},
					},
				},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/products.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "testproducts_gen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := GenerateProductsRegistry(products, tmpDir); err != nil {
		t.Fatalf("TestGenerateProductsRegistry failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmpDir, "products.go"))
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output does not match golden file\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

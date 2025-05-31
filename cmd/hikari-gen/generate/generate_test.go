package generate

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/types"
)

//go:embed testdata
var testdataFS embed.FS

func TestGenerateEnums(t *testing.T) {
	enums := map[string]types.Enum{
		"TestEnum": {
			Type: "uint8",
			Values: []types.EnumValue{
				{Name: "first value", Value: 1},
				{Name: "second-value", Value: 2},
				{Name: "third.value", Value: 3},
			},
		},
	}

	// Read golden file content from embedded FS
	want, err := fs.ReadFile(testdataFS, "testdata/enums.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp file for output
	tmpFile, err := os.CreateTemp("", "enums_*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	if err := generateEnums(tmpFilePath, enums); err != nil {
		t.Fatalf("generateEnums failed: %v", err)
	}

	got, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output does not match golden file\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

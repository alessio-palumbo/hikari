package generate

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
)

//go:embed testdata
var testdataFS embed.FS

func Test_generateEnums(t *testing.T) {
	enums := []decode.Enum{
		{
			Name: "TestEnum",
			Type: "uint8",
			Values: []decode.EnumValue{
				{Name: "first value", Value: 1},
				{Name: "second-value", Value: 2},
				{Name: "third.value", Value: 3},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/enums.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

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

func Test_generateFields(t *testing.T) {
	fields := []decode.FieldGroup{
		{
			Name:      "TestFields",
			SizeBytes: 16,
			Fields: []decode.Field{
				{Name: "Serial", Type: "[6]byte", SizeBytes: 6},
				{Name: "Reserved", Type: "[10]byte", SizeBytes: 10},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/fields.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp file for output
	tmpFile, err := os.CreateTemp("", "fields_*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	if err := generateFields(tmpFilePath, fields); err != nil {
		t.Fatalf("generateFields failed: %v", err)
	}

	got, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output does not match golden file\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

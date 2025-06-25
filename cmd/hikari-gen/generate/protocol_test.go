package generate

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
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

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "testenums_gen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, "enums.go")
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
				{Type: "reserved", SizeBytes: 10},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/fields.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "testfields_gen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, "fields.go")
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

func Test_generateUnions(t *testing.T) {
	unions := []decode.Union{
		{
			Name:      "TestUnion",
			SizeBytes: 16,
			Fields: []decode.Field{
				{Name: "Serial", Type: "[6]byte", SizeBytes: 6},
				{Type: "reserved", SizeBytes: 10},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/unions.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "testunions_gen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, "unions.go")
	if err := generateUnions(tmpFilePath, unions); err != nil {
		t.Fatalf("generateUnions failed: %v", err)
	}

	got, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output does not match golden file\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func Test_generatePackets(t *testing.T) {
	packets := []decode.Packet{
		{
			Name:      "TestPacket",
			Namespace: "testpacket",
			SizeBytes: 16,
			Fields: []decode.Field{
				{Name: "Serial", Type: "[6]byte", SizeBytes: 6},
				{Type: "reserved", SizeBytes: 10},
			},
		},
	}

	want, err := fs.ReadFile(testdataFS, "testdata/packets.go")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "testpacket_gen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := generatePackets(tmpDir, packets); err != nil {
		t.Fatalf("generatePackets failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmpDir, "testpacket.go"))
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output does not match golden file\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

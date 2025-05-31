package generate

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/types"
)

//go:embed templates/*
var templates embed.FS

// Generate runs all generation steps.
func Generate(spec types.ProtocolSpec, outputRoot string) error {
	if err := os.MkdirAll(filepath.Join(outputRoot, "types"), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := generateEnums(filepath.Join(outputRoot, "types", "enums.go"), spec.Enums); err != nil {
		return fmt.Errorf("generating enums: %w", err)
	}

	// Add other generateXYZ() calls here.

	return nil
}

func generateEnums(outputPath string, enums map[string]types.Enum) error {
	tmplBytes, err := templates.ReadFile("templates/enums.tmpl")
	if err != nil {
		return fmt.Errorf("reading enums template: %w", err)
	}

	tmpl, err := template.New("enums").Funcs(template.FuncMap{
		"camelcase": camelcase,
	}).Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, enums); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	// Format with gofmt
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(formatted); err != nil {
		return fmt.Errorf("writing formatted output: %w", err)
	}
	return nil
}

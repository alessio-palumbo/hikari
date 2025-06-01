package generate

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
)

//go:embed templates/*
var templates embed.FS

// Generate runs all generation steps.
func Generate(spec *decode.ProtocolSpec, outputRoot string) error {
	if err := os.MkdirAll(filepath.Join(outputRoot, "types"), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := generateEnums(filepath.Join(outputRoot, "types", "enums.go"), spec.Enums); err != nil {
		return fmt.Errorf("generating enums: %w", err)
	}
	if err := generateFields(filepath.Join(outputRoot, "types", "fields.go"), spec.Fields); err != nil {
		return fmt.Errorf("generating fields: %w", err)
	}

	// Add other generateXYZ() calls here.

	return nil
}

func generateEnums(outputPath string, enums []decode.Enum) error {
	filtered := make(map[string]decode.Enum)
	for _, enum := range enums {
		var values []decode.EnumValue
		for _, v := range enum.Values {
			if strings.ToLower(v.Name) != "reserved" {
				values = append(values, v)
			}
		}
		if len(values) > 0 {
			filtered[enum.Name] = decode.Enum{
				Name:   enum.Name,
				Values: values,
			}
		}
	}
	return generateFromTemplate("templates/enums.tmpl", outputPath, enums)
}

func generateFields(outputPath string, fields []decode.FieldGroup) error {
	return generateFromTemplate("templates/fields.tmpl", outputPath, fields)
}

func generateFromTemplate(tmplPath, outputPath string, data any) error {
	tmplBytes, err := templates.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", tmplPath, err)
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).Funcs(template.FuncMap{
		"camelcase": camelcase,
	}).Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

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

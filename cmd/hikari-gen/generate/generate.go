package generate

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
	"golang.org/x/tools/imports"
)

//go:embed templates/*
var templates embed.FS

// Generate runs all generation steps.
func Generate(spec *decode.ProtocolSpec, outputRoot string) error {
	if err := generateEnums(filepath.Join(outputRoot, "enums", "enums.go"), spec.Enums); err != nil {
		return fmt.Errorf("generating enums: %w", err)
	}
	if err := generateFields(filepath.Join(outputRoot, "packets", "fields.go"), spec.Fields); err != nil {
		return fmt.Errorf("generating fields: %w", err)
	}
	if err := generateUnions(filepath.Join(outputRoot, "packets", "unions.go"), spec.Unions); err != nil {
		return fmt.Errorf("generating unions: %w", err)
	}
	if err := generatePayloadTypes(filepath.Join(outputRoot, "packets", "payloads.go"), spec.Packets); err != nil {
		return fmt.Errorf("generating payloads: %w", err)
	}
	if err := generatePackets(filepath.Join(outputRoot, "packets"), spec.Packets); err != nil {
		return fmt.Errorf("generating packets: %w", err)
	}

	return nil
}

func generateEnums(outputPath string, enums []decode.Enum) error {
	var filtered []decode.Enum
	for _, enum := range enums {
		var values []decode.EnumValue
		for _, v := range enum.Values {
			if strings.ToLower(v.Name) != "reserved" {
				values = append(values, v)
			}
		}
		if len(values) > 0 {
			filtered = append(filtered, decode.Enum{
				Name:   enum.Name,
				Type:   enum.Type,
				Values: values,
			})
		}
	}
	return generateFromTemplate("templates/enums.tmpl", outputPath, filtered)
}

func generateFields(outputPath string, fields []decode.FieldGroup) error {
	for _, f := range fields {
		fixReservedFieldNames(f.Fields)
	}
	return generateFromTemplate("templates/fields.tmpl", outputPath, fields)
}

func generateUnions(outputPath string, unions []decode.Union) error {
	for _, u := range unions {
		var fields []decode.Field
		for _, f := range u.Fields {
			if f.Type != "reserved" {
				fields = append(fields, f)
			}
		}
		u.Fields = fields
	}
	return generateFromTemplate("templates/unions.tmpl", outputPath, unions)
}

func generatePayloadTypes(outputPath string, packets []decode.Packet) error {
	sorted := slices.SortedFunc(slices.Values(packets), func(a, b decode.Packet) int {
		return a.PktType - b.PktType
	})
	return generateFromTemplate("templates/payloads.tmpl", outputPath, sorted)
}

func generatePackets(outputPath string, packets []decode.Packet) error {
	var namespaces []string
	nsMap := make(map[string][]decode.Packet)

	for _, f := range packets {
		if _, ok := nsMap[f.Namespace]; !ok {
			namespaces = append(namespaces, f.Namespace)
		}
		fixReservedFieldNames(f.Fields)
		nsMap[f.Namespace] = append(nsMap[f.Namespace], f)
	}

	if err := generateFromTemplate("templates/helpers.tmpl", filepath.Join(outputPath, "helpers.go"), nil); err != nil {
		return err
	}

	for _, ns := range namespaces {
		if err := generateFromTemplate("templates/packets.tmpl", filepath.Join(outputPath, ns+".go"), nsMap[ns]); err != nil {
			return err
		}
	}
	return nil
}

func generateFromTemplate(tmplPath, outputPath string, data any) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory [%s]: %w", dir, err)
	}

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

	formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("goimports failed [%s]: %w", tmplPath, err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file [%s]: %w", outputPath, err)
	}
	defer f.Close()

	if _, err := f.Write(formatted); err != nil {
		return fmt.Errorf("writing formatted output: %w", err)
	}

	return nil
}

func fixReservedFieldNames(fields []decode.Field) {
	reservedCount := 0
	for i := range fields {
		f := &fields[i]
		if f.Type == "reserved" {
			reservedCount++
			f.Name = fmt.Sprintf("Reserved%d", reservedCount)
			f.Type = reserveTypeForSizeBytes(f.SizeBytes)
		}
	}
}

func reserveTypeForSizeBytes(size int) string {
	switch size {
	case 1:
		return "uint8"
	case 2:
		return "uint16"
	case 4:
		return "uint32"
	case 8:
		return "uint64"
	default:
		return fmt.Sprintf("[%-d]byte", size)
	}
}

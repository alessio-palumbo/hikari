package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/types"
	"gopkg.in/yaml.v3"
)

//go:embed src/protocol.yml
var protocolYAML []byte

func main() {
	var spec types.ProtocolSpec
	if err := yaml.Unmarshal(protocolYAML, &spec); err != nil {
		log.Fatalf("Failed to parse embedded protocol.yml: %v", err)
	}

	if err := ensureDirs(); err != nil {
		log.Fatalf("Failed to create output directories: %v", err)
	}

	// Placeholder
	fmt.Println("Parsed enums:")
	for name := range spec.Enums {
		fmt.Println(" -", name)
	}

	fmt.Println("Parsed field groups:")
	for name := range spec.Fields {
		fmt.Println(" -", name)
	}

	fmt.Println("Parsed packets:")
	for name := range spec.Packets {
		fmt.Println(" -", name)
	}

	// Code generation logic to follow...
}

func ensureDirs() error {
	return os.MkdirAll("gen/protocol/types", 0755)
}

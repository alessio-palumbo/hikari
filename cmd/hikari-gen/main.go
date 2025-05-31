package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/generate"
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

	if err := generate.Generate(spec, "gen/protocol"); err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}

	fmt.Println("Code generation completed successfully.")
}

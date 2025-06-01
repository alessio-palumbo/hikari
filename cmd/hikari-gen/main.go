package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/generate"
)

const (
	generateDir = "gen/protocol"
)

//go:embed src/protocol.yml
var protocolYAML []byte

func main() {
	spec, err := decode.DecodeProtocol(protocolYAML)
	if err != nil {
		log.Fatalf("Failed to parse embedded protocol.yml: %v", err)
	}

	if err := generate.Generate(spec, generateDir); err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}

	fmt.Println("Code generation completed successfully.")
}

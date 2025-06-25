package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/generate"
)

const (
	protocolGenerateDir = "gen/protocol"
	registryGenerateDir = "gen/registry"
)

//go:embed src/protocol.yml
var protocolYAML []byte

//go:embed src/products.json
var productsJSON []byte

func main() {
	protocolSpec, err := decode.DecodeProtocol(protocolYAML)
	if err != nil {
		log.Fatalf("Failed to parse embedded protocol.yml: %v", err)
	}

	if err := generate.GenerateProtocol(protocolSpec, protocolGenerateDir); err != nil {
		log.Fatalf("Failed to generate Protocol: %v", err)
	}

	productsSpec, err := decode.DecodeProductsRegistry(productsJSON)
	if err != nil {
		log.Fatalf("Failed to parse embedded products.json: %v", err)
	}

	if err := generate.GenerateProductsRegistry(productsSpec, registryGenerateDir); err != nil {
		log.Fatalf("Failed to generate Products registry: %v", err)
	}

	fmt.Println("Code generation completed successfully.")
}

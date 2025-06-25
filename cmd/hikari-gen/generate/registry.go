package generate

import (
	"path/filepath"

	"github.com/alessio-palumbo/hikari/cmd/hikari-gen/decode"
)

func GenerateProductsRegistry(products []decode.Product, outputRoot string) error {
	if err := generateFromTemplate("templates/products.tmpl", filepath.Join(outputRoot, "products.go"), products); err != nil {
		return err
	}
	return nil
}

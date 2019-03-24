package infrastructure

import (
	"context"

	"flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	"flamingo.me/flamingo-commerce/v3/product/domain"
)

type (
	// ProductServiceAdapter interface
	ProductServiceAdapter struct {
		InMemoryProductRepository *infrastructure.InMemoryProductRepository `inject:""`
	}
)

var (
	brands = []string{
		"Apple",
		"Bose",
		"Dior",
		"Hugo Boss",
	}
)

// Get returns a product struct
func (ps *ProductServiceAdapter) Get(ctx context.Context, marketplaceCode string) (domain.BasicProduct, error) {
	return ps.InMemoryProductRepository.FindByMarketplaceCode(marketplaceCode)
}

package infrastructure

import (
	"context"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	"flamingo.me/flamingo-commerce/product/domain"
)

type (
	// ProductService interface
	ProductServiceAdapter struct {
		InMemoryProductRepository *productRepository.InMemoryProductRepository `inject:""`
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

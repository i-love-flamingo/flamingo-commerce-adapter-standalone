package infrastructure

import (
	"flamingo.me/flamingo-commerce/product/domain"
)

type (
	InMemoryProductRepository struct {
		products map[string]domain.BasicProduct
	}
)

// Get returns a product struct
func (r *InMemoryProductRepository) Add(product domain.BasicProduct) error {
	if r.products == nil {
		r.products = make(map[string]domain.BasicProduct)
	}
	r.products[product.BaseData().MarketPlaceCode] = product
	return nil
}

// Get returns a product struct
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	return nil, nil
}

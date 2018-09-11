package infrastructure

import (
	"flamingo.me/flamingo-commerce/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/search/domain"
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
	if v, ok := r.products[marketplaceCode]; ok {
		return v, nil
	}
	return nil, domain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

// Get returns a product struct
func (r *InMemoryProductRepository) Find(filter ...searchDomain.Filter) ([]domain.BasicProduct, error) {
	var results []domain.BasicProduct
	for _, p := range r.products {
		results = append(results, p)
	}
	return results, nil
}

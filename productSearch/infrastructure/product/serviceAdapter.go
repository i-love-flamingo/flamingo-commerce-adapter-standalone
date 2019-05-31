package product

import (
	"context"

	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"
	"flamingo.me/flamingo-commerce/v3/product/domain"
)

type (
	// ServiceAdapter - implements flamingo_commerce.ProductService interface
	ServiceAdapter struct {
		productRepository productSearch.ProductRepository
	}
)


//Inject - dingo injector
func (ps *ServiceAdapter) Inject(productRepository productSearch.ProductRepository)  {
	ps.productRepository = productRepository
}


// Get returns a product struct
func (ps *ServiceAdapter) Get(ctx context.Context, marketplaceCode string) (domain.BasicProduct, error) {
	return ps.productRepository.FindByMarketplaceCode(marketplaceCode)
}

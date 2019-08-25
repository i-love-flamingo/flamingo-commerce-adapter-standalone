package product

import (
	"context"
	domain "flamingo.me/flamingo-commerce-adapter-standalone/productSearch/domain"

	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
)

type (
	// ServiceAdapter - implements flamingo_commerce.ProductService interface
	ServiceAdapter struct {
		productRepository domain.ProductRepository
	}
)


//Inject - dingo injector
func (ps *ServiceAdapter) Inject(productRepository domain.ProductRepository)  {
	ps.productRepository = productRepository
}


// Get returns a product struct
func (ps *ServiceAdapter) Get(ctx context.Context, marketplaceCode string) (productDomain.BasicProduct, error) {
	return ps.productRepository.FindByMarketplaceCode(marketplaceCode)
}

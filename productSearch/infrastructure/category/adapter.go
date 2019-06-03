package category

import (
	"context"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
)

type (
	//Adapter - Adapter that uses a category csv (WIP)
	Adapter struct {
		productRepository productSearch.ProductRepository
	}
)



var (
	_ categoryDomain.CategoryService = &Adapter{}
)

//Inject - dingo injector
func (ps *Adapter) Inject(productRepository productSearch.ProductRepository)  {
	ps.productRepository = productRepository
}

// Tree a category
func (c *Adapter) Tree(ctx context.Context, activeCategoryCode string) (categoryDomain.Tree, error) {
	return c.productRepository.CategoryTree(activeCategoryCode)
}


// Get a category with more data
func (c *Adapter) Get(ctx context.Context, categoryCode string) (categoryDomain.Category, error) {
	return c.productRepository.Category(categoryCode)
}

package infrastructure

import (
	"context"

	"flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
)

type (
	//CategoryServiceAdapter - Adapter that uses a category csv (WIP)
	CategoryServiceAdapter struct {
		inMemoryProductRepository *infrastructure.InMemoryProductRepository `inject:""`
	}
)

var (
	_ categoryDomain.CategoryService = new(CategoryServiceAdapter)
)

func (c *CategoryServiceAdapter) Inject(inMemoryProductRepository *infrastructure.InMemoryProductRepository) {
	c.inMemoryProductRepository = inMemoryProductRepository

}

// Tree a category
func (c *CategoryServiceAdapter) Tree(ctx context.Context, activeCategoryCode string) (categoryDomain.Tree, error) {
	return nil, nil
}

// Get a category with more data
func (c *CategoryServiceAdapter) Get(ctx context.Context, categoryCode string) (categoryDomain.Category, error) {
	return nil, nil
}

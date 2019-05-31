package category

import (
	"context"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
)

type (
	//Adapter - Adapter that uses a category csv (WIP)
	Adapter struct {
	}
)

var (
	_ categoryDomain.CategoryService = &Adapter{}
)

func (c *Adapter) Inject() {


}

// Tree a category
func (c *Adapter) Tree(ctx context.Context, activeCategoryCode string) (categoryDomain.Tree, error) {
	return nil, nil
}

// Get a category with more data
func (c *Adapter) Get(ctx context.Context, categoryCode string) (categoryDomain.Category, error) {
	return nil, nil
}

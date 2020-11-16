package category

import (
	"context"

	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
)

type (
	// Adapter that uses the category repository
	Adapter struct {
		categoryRepository domain.CategoryRepository
		logger             flamingo.Logger
	}
)

var (
	_ categoryDomain.CategoryService = &Adapter{}
)

// Inject dependencies
func (a *Adapter) Inject(categoryRepository domain.CategoryRepository, logger flamingo.Logger) {
	a.categoryRepository = categoryRepository
	a.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone.commercesearch").WithField(flamingo.LogKeyCategory, "categoryadapter")
}

// Tree a category
func (a *Adapter) Tree(ctx context.Context, activeCategoryCode string) (categoryDomain.Tree, error) {
	t, err := a.categoryRepository.CategoryTree(ctx, activeCategoryCode)
	if err == categoryDomain.ErrNotFound {
		a.logger.Warn(err)
	}
	if err != nil {
		a.logger.Error(err)
	}
	a.logger.Info("Tree ", t, err)
	return t, err
}

// Get a category with more data
func (a *Adapter) Get(ctx context.Context, categoryCode string) (categoryDomain.Category, error) {
	t, err := a.categoryRepository.Category(ctx, categoryCode)
	if err == categoryDomain.ErrNotFound {
		a.logger.Warn(err)
	}
	if err != nil {
		a.logger.Error(err)
	}
	return t, err
}

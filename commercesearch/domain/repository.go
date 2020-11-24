package domain

import (
	"context"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (
	// ProductRepository port
	ProductRepository interface {
		FindByMarketplaceCode(ctx context.Context, marketplaceCode string) (domain.BasicProduct, error)
		Find(ctx context.Context, filters ...searchDomain.Filter) (*domain.SearchResult, error)
		PrepareIndex(ctx context.Context) error
		UpdateProducts(ctx context.Context, products []domain.BasicProduct) error
		ClearProducts(ctx context.Context, productIds []string) error
		DocumentsCount() int64
	}

	// CategoryRepository port
	CategoryRepository interface {
		PrepareIndex(ctx context.Context) error
		CategoryTree(ctx context.Context, code string) (categoryDomain.Tree, error)
		Category(ctx context.Context, code string) (categoryDomain.Category, error)
		UpdateByCategoryTeasers(ctx context.Context, categories []domain.CategoryTeaser) error
		ClearCategories(ctx context.Context, productIds []string) error
	}
)

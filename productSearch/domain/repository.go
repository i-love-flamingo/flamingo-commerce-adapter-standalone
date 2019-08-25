package domain

import (
	"flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
)

type (
	// ProductRepository - interface
	ProductRepository interface {
		FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error)
		Find(filters ...searchDomain.Filter) (*domain.SearchResult, error)
		CategoryTree(code string) (categoryDomain.Tree, error)
		Category(code string) (categoryDomain.Category, error)

		PrepareIndex() error
		Add(product domain.BasicProduct) error
	}
)

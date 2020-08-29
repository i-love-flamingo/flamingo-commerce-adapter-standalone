package product

import (
	"context"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"

	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (
	// SearchServiceAdapter implements methods to search in a product repository
	SearchServiceAdapter struct {
		productRepository domain.ProductRepository
	}
)

var (
	_ productDomain.SearchService = &SearchServiceAdapter{}
)

//Inject - dingo injector
func (ps *SearchServiceAdapter) Inject(productRepository domain.ProductRepository) {
	ps.productRepository = productRepository
}

// Search returns a Search Result for the given context and supplied filters
func (ps *SearchServiceAdapter) Search(ctx context.Context, filter ...searchDomain.Filter) (*productDomain.SearchResult, error) {
	return ps.productRepository.Find(ctx, filter...)
}

// SearchBy returns Products prefiltered by the given attribute (also based on additional given Filters) e.g. SearchBy(ctx,"brandCode","apple")
func (ps *SearchServiceAdapter) SearchBy(ctx context.Context, attribute string, values []string, filter ...searchDomain.Filter) (*productDomain.SearchResult, error) {
	return ps.Search(ctx, nil)
}

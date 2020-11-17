package product

import (
	"context"

	"github.com/google/go-cmp/cmp"

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

// Inject dependencies
func (ps *SearchServiceAdapter) Inject(productRepository domain.ProductRepository) {
	ps.productRepository = productRepository
}

// Search returns a Search Result for the given context and supplied filters
func (ps *SearchServiceAdapter) Search(ctx context.Context, filter ...searchDomain.Filter) (*productDomain.SearchResult, error) {
	return ps.productRepository.Find(ctx, filter...)
}

// SearchBy returns Products prefiltered by the given attribute (also based on additional given Filters) e.g. SearchBy(ctx, "brandCode", []string{"apple"})
func (ps *SearchServiceAdapter) SearchBy(ctx context.Context, attribute string, values []string, filters ...searchDomain.Filter) (*productDomain.SearchResult, error) {
	attributeFilterPresent := false
	for _, filter := range filters {
		if keyValueFilter, ok := filter.(*searchDomain.KeyValueFilter); ok {
			if keyValueFilter.Key() == attribute && cmp.Equal(keyValueFilter.KeyValues(), values) {
				attributeFilterPresent = true
			}
		}
	}

	if !attributeFilterPresent {
		filters = append(filters, searchDomain.NewKeyValueFilter(attribute, values))
	}

	return ps.Search(ctx, filters...)
}

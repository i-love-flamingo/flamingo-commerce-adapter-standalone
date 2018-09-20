package infrastructure

import (
	"sort"

	"flamingo.me/flamingo-commerce/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/search/domain"
)

type (
	// InMemoryProductRepository serves as a Repository of Products held in memory
	InMemoryProductRepository struct {
		products map[string]domain.BasicProduct
	}
)

// Add returns a product struct
func (r *InMemoryProductRepository) Add(product domain.BasicProduct) error {
	if r.products == nil {
		r.products = make(map[string]domain.BasicProduct)
	}
	r.products[product.BaseData().MarketPlaceCode] = product

	return nil
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	if v, ok := r.products[marketplaceCode]; ok {
		return v, nil
	}
	return nil, domain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *InMemoryProductRepository) Find(filters ...searchDomain.Filter) ([]domain.BasicProduct, error) {
	var results []domain.BasicProduct

	keyValueFilters := getKeyValueFilter(filters...)

	productLoop:
	for _, p := range r.products {
		if len(keyValueFilters) == 0 {
			results = append(results, p)
			continue productLoop
		}

		for _, filter := range keyValueFilters {
			filterKey, values := filter.Value()
			for _, filterVal := range values {
				if !p.BaseData().HasAttribute(filterKey) {
					continue productLoop
				}
				if len(p.BaseData().Attributes[filterKey].Values()) > 0 {
					// Multivalue Attribute
					for _, attributeValue := range p.BaseData().Attributes[filterKey].Values() {
						if attributeValue == filterVal {
							results = append(results, p)
							continue productLoop
						}
					}
				} else if filterVal == p.BaseData().Attributes[filterKey].Value() {
					// Single Value Attribute
					results = append(results, p)
					continue productLoop
				}
			}
		}
	}

	// Sort the Results
	for _, filter := range filters {
		if sortFilter, ok := filter.(*searchDomain.SortFilter); ok {
			k, v := sortFilter.Value()

			sort.Slice(results, func(i, j int) bool {
				if v[0] == "A" {
					return results[i].BaseData().Attributes[k].Value() < results[j].BaseData().Attributes[k].Value()
				}

				return results[i].BaseData().Attributes[k].Value() > results[j].BaseData().Attributes[k].Value()
			})
		}
	}

	// Limit the Results
	for _, filter := range filters {
		if pageSize, ok := filter.(*searchDomain.PaginationPageSize); ok {
			size := pageSize.GetPageSize()
			results = append([]domain.BasicProduct(nil), results[:size]...)
		}
	}

	return results, nil
}

func getKeyValueFilter(filters ...searchDomain.Filter) []*searchDomain.KeyValueFilter {
	var kvFilters []*searchDomain.KeyValueFilter
	for _, filter := range filters {
		// currently only keyvalue filters supported here
		if kvFilter, ok := filter.(*searchDomain.KeyValueFilter); ok {
			kvFilters = append(kvFilters, kvFilter)
		}
	}

	return kvFilters
}

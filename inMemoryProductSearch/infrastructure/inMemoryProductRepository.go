package infrastructure

import (
	"sort"

	"flamingo.me/flamingo-commerce/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/search/domain"
)

type (
	// InMemoryProductRepository serves as a Repository of Products held in memory
	InMemoryProductRepository struct {
		products []domain.BasicProduct
		index map[string]map[string][]*domain.BasicProduct
	}
)

// Add appends a product to the Product Repository
func (r *InMemoryProductRepository) Add(product domain.BasicProduct) error {
	r.products = append(r.products, product)

	if r.index == nil {
		r.index = make(map[string]map[string][]*domain.BasicProduct)
	}

	for _, attribute := range product.BaseData().Attributes {
		if _, ok := r.index[attribute.Code]; !ok {
			r.index[attribute.Code] = make(map[string][]*domain.BasicProduct)
		}

		r.index[attribute.Code][attribute.Value()] = append(r.index[attribute.Code][attribute.Value()], &product)
	}

	return nil
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	results, err := r.Find(searchDomain.NewKeyValueFilter("marketplaceCode", []string{marketplaceCode}))
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, domain.ProductNotFound{
			MarketplaceCode: marketplaceCode,
		}
	}

	return results[0], nil
}

// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *InMemoryProductRepository) Find(filters ...searchDomain.Filter) ([]domain.BasicProduct, error) {
	var results []domain.BasicProduct

	keyValueFilters := getKeyValueFilter(filters...)

	if len(keyValueFilters) == 0 {
		for _, p := range r.products {
			results = append(results, p)
		}
	}

	for _, filter := range keyValueFilters {
		filterKey, filterValues := filter.Value()
		for _, filterValue := range filterValues {
			lookup := r.index[filterKey][filterValue]
			if len(lookup) > 0 {
				for _, productItem := range lookup {
					results = append(results, *productItem)
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
			if len(results) < size {
				size = len(results)
			}

			if size > 0 {
				results = results[:size]
			}
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

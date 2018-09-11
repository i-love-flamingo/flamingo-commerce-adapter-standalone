package productRepository

import (
	"sort"
	"flamingo.me/flamingo-commerce/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/search/domain"
	"flamingo.me/flamingo/framework/flamingo"
)

type (
	InMemoryProductRepository struct {
		products map[string]domain.BasicProduct
	}

	InMemoryProductRepositoryProvider struct {
		InMemoryProductRepositoryFactory *InMemoryProductRepositoryFactory `inject:""`
		Logger                           flamingo.Logger                   `inject:""`
		Locale                           string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.locale"`
		Currency                         string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.currency"`
		ProductCSVPath                   string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
	}
)

//TODO - use map with lock (sync map) https://github.com/golang/go/blob/master/src/sync/map.go
var buildedRepositoryByLocale = make(map[string]*InMemoryProductRepository)

// Get returns a product struct
func (r *InMemoryProductRepository) add(product domain.BasicProduct) error {
	if r.products == nil {
		r.products = make(map[string]domain.BasicProduct)
	}
	r.products[product.BaseData().MarketPlaceCode] = product
	return nil
}

// Get returns a product struct
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	if v, ok := r.products[marketplaceCode]; ok {
		return v, nil
	}
	return nil, domain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

// Get returns a product struct
func (r *InMemoryProductRepository) Find(filters ...searchDomain.Filter) ([]domain.BasicProduct, error) {
	var results []domain.BasicProduct

productLoop:
	for _, p := range r.products {
		for _, filter := range filters {
			// currently only keyvalue filters supported here
			if _, ok := filter.(*searchDomain.KeyValueFilter); ok {
				filterKey, values := filter.Value()
				for _, filterVal := range values {
					if p.BaseData().HasAttribute(filterKey) {
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
		}
	}

	// Sort the Results
	for _, filter := range filters {
		if sortFilter, ok := filter.(*searchDomain.SortFilter); ok {
			k, v := sortFilter.Value()

			sort.Slice(results, func(i, j int) bool {
				var result bool

				if v[0] == "A" {
					result = results[i].BaseData().Attributes[k].Value() < results[j].BaseData().Attributes[k].Value()
				}
				if v[0]  == "D" {
					result = results[i].BaseData().Attributes[k].Value() > results[j].BaseData().Attributes[k].Value()
				}

				return result
			})
		}
	}

	return results, nil
}

func (p *InMemoryProductRepositoryProvider) GetForCurrentLocale() (*InMemoryProductRepository, error) {
	locale := p.Locale
	if v, ok := buildedRepositoryByLocale[locale]; ok {
		return v, nil
	}
	p.Logger.Info("Build InMemoryProductRepository for locale " + locale + " .....")
	rep, err := p.InMemoryProductRepositoryFactory.BuildFromProductCSV(p.ProductCSVPath, locale, p.Currency)
	if err != nil {
		return nil, err
	}
	buildedRepositoryByLocale[locale] = rep
	return buildedRepositoryByLocale[locale], nil
}

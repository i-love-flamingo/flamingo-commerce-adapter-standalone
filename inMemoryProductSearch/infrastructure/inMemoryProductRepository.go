package infrastructure

import (
	"log"
	"sort"
	"sync"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"

	"flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (
	// InMemoryProductRepository serves as a Repository of Products held in memory
	InMemoryProductRepository struct {
		//marketplaceCodeIndex - index to get products from marketplaceCode
		marketplaceCodeIndex map[string]domain.BasicProduct

		//attributeReverseIndex - index to get products from attribute
		attributeReverseIndex map[string]map[string][]string

		//categoriesReverseIndex - index to get products by categoryCode
		categoriesReverseIndex map[string][]string
		addReadMutex           sync.RWMutex
	}

	marketPlaceCodeSet struct {
		currentSet    []string
		initialFilled bool
	}
)

// Add appends a product to the Product Repository
func (r *InMemoryProductRepository) Add(product domain.BasicProduct) error {
	r.addReadMutex.Lock()
	defer r.addReadMutex.Unlock()

	marketPlaceCode := product.BaseData().MarketPlaceCode
	//Set reverseindex for marketplaceCode (the primary indendifier)
	if r.marketplaceCodeIndex == nil {
		r.marketplaceCodeIndex = make(map[string]domain.BasicProduct)
	}
	if r.marketplaceCodeIndex[marketPlaceCode] != nil {
		log.Println("Duplicate for marketplace code " + marketPlaceCode)
	}
	r.marketplaceCodeIndex[product.BaseData().MarketPlaceCode] = product

	//Now add product to category indexes:
	if r.categoriesReverseIndex == nil {
		r.categoriesReverseIndex = make(map[string][]string)
	}
	for _, categoryCode := range product.BaseData().CategoryCodes {
		r.categoriesReverseIndex[categoryCode] = append(r.categoriesReverseIndex[categoryCode], marketPlaceCode)
	}

	//Now fill the reverse index for all products attributes:
	if r.attributeReverseIndex == nil {
		r.attributeReverseIndex = make(map[string]map[string][]string)
	}
	for _, attribute := range product.BaseData().Attributes {
		if _, ok := r.attributeReverseIndex[attribute.Code]; !ok {
			r.attributeReverseIndex[attribute.Code] = make(map[string][]string)
		}
		r.attributeReverseIndex[attribute.Code][attribute.Value()] = append(r.attributeReverseIndex[attribute.Code][attribute.Value()], marketPlaceCode)
	}

	return nil
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()

	if product, ok := r.marketplaceCodeIndex[marketplaceCode]; ok {
		return product, nil
	}

	return nil, domain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}

}

// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *InMemoryProductRepository) Find(filters ...searchDomain.Filter) ([]domain.BasicProduct, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()

	var productResults []domain.BasicProduct

	var matchingMarketplaceCodes marketPlaceCodeSet

	for _, filter := range filters {
		filterKey, filterValues := filter.Value()
		switch filter.(type) {
		case *searchDomain.KeyValueFilter:
			for _, filterValue := range filterValues {
				matchingCodes := r.attributeReverseIndex[filterKey][filterValue]
				matchingMarketplaceCodes.intersection(matchingCodes)
			}
		case categoryDomain.CategoryFacet:
			for _, filterValue := range filterValues {
				matchingCodes := r.categoriesReverseIndex[filterValue]
				matchingMarketplaceCodes.intersection(matchingCodes)
			}
		}

	}

	if !matchingMarketplaceCodes.initialFilled {
		//get all products if not filtered yet
		for _, p := range r.marketplaceCodeIndex {
			productResults = append(productResults, p)
		}
	} else {
		//otherwise get only the remaining marketplacecodes
		productResults = r.getMatchingProducts(matchingMarketplaceCodes.currentSet)
	}

	// Sort the Results
	for _, filter := range filters {
		if sortFilter, ok := filter.(*searchDomain.SortFilter); ok {
			k, v := sortFilter.Value()

			sort.Slice(productResults, func(i, j int) bool {
				if v[0] == "A" {
					return productResults[i].BaseData().Attributes[k].Value() < productResults[j].BaseData().Attributes[k].Value()
				}

				return productResults[i].BaseData().Attributes[k].Value() > productResults[j].BaseData().Attributes[k].Value()
			})
		}
	}

	// Limit the Results
	for _, filter := range filters {
		if pageSize, ok := filter.(*searchDomain.PaginationPageSize); ok {
			size := pageSize.GetPageSize()
			if len(productResults) < size {
				size = len(productResults)
			}

			if size > 0 {
				productResults = productResults[:size]
			}
		}
	}

	return productResults, nil
}

func (r *InMemoryProductRepository) getMatchingProducts(codes []string) []domain.BasicProduct {
	var matches []domain.BasicProduct
	for code, product := range r.marketplaceCodeIndex {
		for _, codeToFind := range codes {
			if code == codeToFind {
				matches = append(matches, product)
			}
		}
	}
	return matches
}

func (s *marketPlaceCodeSet) intersection(set2 []string) {
	if !s.initialFilled {
		s.currentSet = set2
		s.initialFilled = true
		return
	}
	var result []string
	for _, v1 := range s.currentSet {
		for _, v2 := range set2 {
			if v1 == v2 {
				result = append(result, v2)
			}
		}
	}
	s.currentSet = result
}

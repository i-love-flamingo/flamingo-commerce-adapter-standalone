package productSearch

import (
	"errors"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"log"
	"math"
	"sort"
	"sync"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"

	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (

	// InMemoryProductRepository serves as a Repository of Products held in memory
	InMemoryProductRepository struct {
		//marketplaceCodeIndex - index to get products from marketplaceCode
		marketplaceCodeIndex map[string]productDomain.BasicProduct

		//attributeReverseIndex - index to get products from attribute
		attributeReverseIndex map[string]map[string][]string

		//productsByCategoriesReverseIndex - index to get products by categoryCode
		productsByCategoriesReverseIndex map[string][]string
		addReadMutex                     sync.RWMutex

		//for category adapters:
		rootCategory *categoryDomain.TreeData
		categorTreeIndex map[string]*categoryDomain.TreeData

		logger flamingo.Logger
	}

	marketPlaceCodeSet struct {
		currentSet    []string
		initialFilled bool
	}
)

var (
	_ domain.ProductRepository = &InMemoryProductRepository{}
)


func (r *InMemoryProductRepository) PrepareIndex() error {
	return nil
}



func (r *InMemoryProductRepository) Inject(logger flamingo.Logger)  {
	r.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone").WithField(flamingo.LogKeyCategory, "InMemoryProductRepository")
}

// Add appends a product to the Product Repository
func (r *InMemoryProductRepository) Add(product productDomain.BasicProduct) error {
	r.addReadMutex.Lock()
	defer r.addReadMutex.Unlock()

	if (product.BaseData().MarketPlaceCode == "" ) {
		log.Println("No marketplace code")
		return errors.New("No marketplace code ")
	}
	marketPlaceCode := product.BaseData().MarketPlaceCode
	//Set reverseindex for marketplaceCode (the primary indendifier)
	if r.marketplaceCodeIndex == nil {
		r.marketplaceCodeIndex = make(map[string]productDomain.BasicProduct)
	}
	if r.marketplaceCodeIndex[marketPlaceCode] != nil {
		err := errors.New("Duplicate for marketplace code " + marketPlaceCode)
		r.logger.Error(err)
		return err
	}
	r.marketplaceCodeIndex[product.BaseData().MarketPlaceCode] = product

	//Now add product to category indexes:
	if r.productsByCategoriesReverseIndex == nil {
		r.productsByCategoriesReverseIndex = make(map[string][]string)
	}


	for _, categoryTeaser := range product.BaseData().Categories {
		r.productsByCategoriesReverseIndex[categoryTeaser.Code] = append(r.productsByCategoriesReverseIndex[categoryTeaser.Code], marketPlaceCode)
		categoryPathForMergeIn := r.categoryTeaserToCategoryTree(categoryTeaser, nil)
		r.rootCategory = r.addCategoryPath(r.rootCategory,categoryPathForMergeIn)
	}
	if product.BaseData().MainCategory.Code != "" {
		r.productsByCategoriesReverseIndex[product.BaseData().MainCategory.Code] = append(r.productsByCategoriesReverseIndex[product.BaseData().MainCategory.Code], marketPlaceCode)
		categoryPathForMergeIn := r.categoryTeaserToCategoryTree(product.BaseData().MainCategory, nil)
		r.rootCategory  = r.addCategoryPath(r.rootCategory,categoryPathForMergeIn)
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
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (productDomain.BasicProduct, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()
	if product, ok := r.marketplaceCodeIndex[marketplaceCode]; ok {
		return product, nil
	}
	return nil, productDomain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

func (r *InMemoryProductRepository) CategoryTree(code string) (categoryDomain.Tree, error)  {
	if r.rootCategory == nil {
		return nil,  errors.New("not found")
	}
	if code == "" {
		return r.rootCategory, nil
	}
	if tree, ok := r.categorTreeIndex[code]; ok {
		return tree, nil
	}
	return nil, errors.New("not found")
}


func (r *InMemoryProductRepository) Category(code string) (categoryDomain.Category, error)  {
	if r.rootCategory == nil {
		return nil,  errors.New("not found")
	}
	if code == "" {
		return &categoryDomain.CategoryData{
			CategoryCode: r.rootCategory.CategoryCode,
			CategoryName: r.rootCategory.CategoryName,
		}, nil
	}
	if tree, ok := r.categorTreeIndex[code]; ok {
		return &categoryDomain.CategoryData{
			CategoryCode: tree.CategoryCode,
			CategoryName: tree.CategoryName,
		}, nil
	}
	return nil, errors.New("not found")
}


// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *InMemoryProductRepository) Find(filters ...searchDomain.Filter) (*productDomain.SearchResult, error) {

	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()

	var productResults []productDomain.BasicProduct

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
				matchingCodes := r.productsByCategoriesReverseIndex[filterValue]
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

	totalHits := len(productResults)
	pageSize := int(0)
	pageAmount := int(0)

	// Limit the Results
	for _, filter := range filters {
		if pageSizeFilter, ok := filter.(*searchDomain.PaginationPageSize); ok {
			size := pageSizeFilter.GetPageSize()
			pageSize = size
			if len(productResults) < size {
				size = len(productResults)
			}

			if size > 0 {
				productResults = productResults[:size]
			}
		}
	}
	if pageSize > 0 {
		pageAmount = int(math.Ceil(float64(totalHits) / float64(pageSize)))
	}

	return &productDomain.SearchResult{
		Hits: productResults,
		Result: searchDomain.Result{
			SearchMeta: searchDomain.SearchMeta{
				NumResults: totalHits,
				NumPages:   pageAmount,
			}},
	}, nil
}

func (r *InMemoryProductRepository) getMatchingProducts(codes []string) []productDomain.BasicProduct {
	var matches []productDomain.BasicProduct
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


//addCategoryPath - merges in the given categoryToAdd to the  passed currentExisting
// also adds new categories to the reverse index
func (r *InMemoryProductRepository) addCategoryPath(currentExisting *categoryDomain.TreeData, categoryToAdd *categoryDomain.TreeData) *categoryDomain.TreeData {
	if r.categorTreeIndex == nil {
		r.categorTreeIndex = make(map[string]*categoryDomain.TreeData)
	}
	if currentExisting == nil {
		clone := *categoryToAdd
		currentExisting = &clone
		currentExisting.SubTreesData = nil
		r.categorTreeIndex[currentExisting.CategoryCode] = currentExisting
	}
	if currentExisting.CategoryCode != categoryToAdd.CategoryCode {
		return currentExisting
	}
	for _,subTreeToAddTree := range categoryToAdd.SubTreesData {
		exists := false
		for _,existingSubTree := range currentExisting.SubTreesData {
			if existingSubTree.CategoryCode == subTreeToAddTree.CategoryCode {
				exists = true
				existingSubTree = r.addCategoryPath(existingSubTree,subTreeToAddTree)
			}
		}
		if !exists {
			currentExisting.SubTreesData = append(currentExisting.SubTreesData,subTreeToAddTree)
			r.categorTreeIndex[subTreeToAddTree.CategoryCode] = subTreeToAddTree
		}
	}
	return currentExisting
}

/**
  sub_sub --parent--> sub --parent--> root


  Passed is "sub_sub"

  Returned is "root"

 */
func (r *InMemoryProductRepository) categoryTeaserToCategoryTree(teaser productDomain.CategoryTeaser, child *categoryDomain.TreeData) *categoryDomain.TreeData {
	if teaser.Code == "" {
		return nil
	}
	var childs []*categoryDomain.TreeData
	if child != nil {
		childs = []*categoryDomain.TreeData{child}
	}
	currentCategory := &categoryDomain.TreeData{
		IsActive:true,
		CategoryName: teaser.Name,
		CategoryCode: teaser.Code,
		SubTreesData: childs,
	}

	if teaser.Parent == nil {
		return currentCategory
	}

	return r.categoryTeaserToCategoryTree(*teaser.Parent,currentCategory)

}
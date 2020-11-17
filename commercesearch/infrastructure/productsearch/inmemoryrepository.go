package productsearch

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"

	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (

	// InMemoryProductRepository serves as a Repository of Products held in memory
	InMemoryProductRepository struct {
		// marketplaceCodeIndex index to get products from marketplaceCode, e.g. get product for market place code 'foobar'
		marketplaceCodeIndex map[string]productDomain.BasicProduct

		// attributeReverseIndex index to get all market place codes for a certain attribute, e.g. all market place codes with attribute 'size' and value 'large'
		attributeReverseIndex map[string]map[string][]string

		// productsByCategoriesReverseIndex index to get all market place codes for a categoryCode, e.g. all market place codes with category 'clothing'
		productsByCategoriesReverseIndex map[string][]string
		addReadMutex                     sync.RWMutex

		// for category adapters:
		rootCategory      *categoryDomain.TreeData
		categoryTreeIndex map[string]*categoryDomain.TreeData

		logger flamingo.Logger
	}

	marketPlaceCodeSet struct {
		currentSet    []string
		initialFilled bool
	}
)

var (
	_ domain.ProductRepository  = &InMemoryProductRepository{}
	_ domain.CategoryRepository = &InMemoryProductRepository{}
)

// PrepareIndex implementation
func (r *InMemoryProductRepository) PrepareIndex(_ context.Context) error {
	return nil
}

// Inject dependencies
func (r *InMemoryProductRepository) Inject(logger flamingo.Logger) {
	r.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone").WithField(flamingo.LogKeyCategory, "InMemoryProductRepository")
}

// UpdateByCategoryTeasers updates or appends a category to the Product Repository
func (r *InMemoryProductRepository) UpdateByCategoryTeasers(_ context.Context, categoryTeasers []productDomain.CategoryTeaser) error {
	for _, categoryTeaser := range categoryTeasers {

		categoryPathForMergeIn := r.categoryTeaserToCategoryTree(categoryTeaser, nil)
		r.rootCategory = r.addCategoryPath(r.rootCategory, categoryPathForMergeIn)
	}

	return nil
}

func printTree(tree categoryDomain.Tree, indend string) {
	fmt.Printf("\n %v > %v", indend, tree.Code())
	for _, s := range tree.SubTrees() {
		printTree(s, indend+"   ")
	}
}

// ClearCategories clears given ids from repo
func (r *InMemoryProductRepository) ClearCategories(_ context.Context, categoryIds []string) error {
	return nil
}

// ClearProducts from Product Repository
func (r *InMemoryProductRepository) ClearProducts(_ context.Context, products []string) error {
	return nil
}

// UpdateProducts (add or update) to the Product Repository
func (r *InMemoryProductRepository) UpdateProducts(_ context.Context, products []productDomain.BasicProduct) error {
	r.addReadMutex.Lock()
	defer r.addReadMutex.Unlock()

	for _, product := range products {
		if product.BaseData().MarketPlaceCode == "" {
			return errors.New("No marketplace code ")
		}
		marketPlaceCode := product.BaseData().MarketPlaceCode

		err := r.addProductToMarketplaceCodeReverseIndex(marketPlaceCode, product)
		if err != nil {
			return err
		}

		r.addMarketplaceCodeToCategoryReverseIndex(product, marketPlaceCode)
		r.addMarketplaceCodeToAttributeReverseIndex(product, marketPlaceCode)
	}

	return nil
}

func (r *InMemoryProductRepository) addMarketplaceCodeToAttributeReverseIndex(product productDomain.BasicProduct, marketPlaceCode string) {
	if r.attributeReverseIndex == nil {
		r.attributeReverseIndex = make(map[string]map[string][]string)
	}
	for _, attribute := range product.BaseData().Attributes {
		if _, ok := r.attributeReverseIndex[attribute.Code]; !ok {
			r.attributeReverseIndex[attribute.Code] = make(map[string][]string)
		}
		r.attributeReverseIndex[attribute.Code][attribute.Value()] = append(r.attributeReverseIndex[attribute.Code][attribute.Value()], marketPlaceCode)
	}
}

func (r *InMemoryProductRepository) addMarketplaceCodeToCategoryReverseIndex(product productDomain.BasicProduct, marketPlaceCode string) {
	if r.productsByCategoriesReverseIndex == nil {
		r.productsByCategoriesReverseIndex = make(map[string][]string)
	}

	for _, categoryTeaser := range product.BaseData().Categories {
		r.productsByCategoriesReverseIndex[categoryTeaser.Code] = append(r.productsByCategoriesReverseIndex[categoryTeaser.Code], marketPlaceCode)
	}
	if product.BaseData().MainCategory.Code != "" {
		if !inSlice(r.productsByCategoriesReverseIndex[product.BaseData().MainCategory.Code], marketPlaceCode) {
			r.productsByCategoriesReverseIndex[product.BaseData().MainCategory.Code] = append(r.productsByCategoriesReverseIndex[product.BaseData().MainCategory.Code], marketPlaceCode)
		}
	}
}

func (r *InMemoryProductRepository) addProductToMarketplaceCodeReverseIndex(marketPlaceCode string, product productDomain.BasicProduct) error {
	// Set reverse index for marketplaceCode (the primary identifier)
	if r.marketplaceCodeIndex == nil {
		r.marketplaceCodeIndex = make(map[string]productDomain.BasicProduct)
	}
	if r.marketplaceCodeIndex[marketPlaceCode] != nil {
		err := errors.New("Duplicate for marketplace code " + marketPlaceCode)
		r.logger.Error(err)
		return err
	}
	r.marketplaceCodeIndex[marketPlaceCode] = product
	return nil
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *InMemoryProductRepository) FindByMarketplaceCode(_ context.Context, marketplaceCode string) (productDomain.BasicProduct, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()
	if product, ok := r.marketplaceCodeIndex[marketplaceCode]; ok {
		return product, nil
	}
	return nil, productDomain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

// CategoryTree returns tree - empty code returns RootNode
func (r *InMemoryProductRepository) CategoryTree(_ context.Context, code string) (categoryDomain.Tree, error) {
	if r.rootCategory == nil {
		err := errors.New("category " + code + "not found. No tree indexed")
		return nil, err
	}
	if code == "" {
		return r.rootCategory, nil
	}
	if tree, ok := r.categoryTreeIndex[code]; ok {
		return tree, nil
	}
	return nil, categoryDomain.ErrNotFound
}

// Category returns category - empty code returns root cat
func (r *InMemoryProductRepository) Category(_ context.Context, code string) (categoryDomain.Category, error) {
	if r.rootCategory == nil {
		return nil, errors.New("root not found")
	}
	if code == "" {
		return &categoryDomain.CategoryData{
			CategoryCode: r.rootCategory.CategoryCode,
			CategoryName: r.rootCategory.CategoryName,
		}, nil
	}
	if tree, ok := r.categoryTreeIndex[code]; ok {
		return &categoryDomain.CategoryData{
			CategoryCode: tree.CategoryCode,
			CategoryName: tree.CategoryName,
		}, nil
	}
	return nil, errors.New("not found")
}

// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *InMemoryProductRepository) Find(_ context.Context, filters ...searchDomain.Filter) (*productDomain.SearchResult, error) {

	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()

	var productResults []productDomain.BasicProduct

	var matchingMarketplaceCodes marketPlaceCodeSet

	pageSize := 100
	pageNumber := 1
	sortField := "relevance"
	sortDirection := searchDomain.SortDirectionAscending
	for _, filter := range filters {
		filterKey, filterValues := filter.Value()
		switch f := filter.(type) {
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
		case *searchDomain.PaginationPageSize:
			pageSize = f.GetPageSize()
		case *searchDomain.PaginationPage:
			pageNumber = f.GetPage()
		case *searchDomain.SortFilter:
			sortField = f.Field()
			sortDirection = f.Direction()
		}
	}

	if !matchingMarketplaceCodes.initialFilled {
		// get all products if not filtered yet
		for _, p := range r.marketplaceCodeIndex {
			productResults = append(productResults, p)
		}
	} else {
		// otherwise get only the remaining marketplace codes
		productResults = r.getMatchingProducts(matchingMarketplaceCodes.currentSet)
	}

	// Sort the Results
	sort.Slice(productResults, func(i, j int) bool {
		iV := productResults[i].BaseData().Attributes[sortField].Value()
		jV := productResults[j].BaseData().Attributes[sortField].Value()

		if sortField == "title" || sortField == "relevance" {
			iV = productResults[i].BaseData().Title
			jV = productResults[j].BaseData().Title
		}

		if sortDirection == searchDomain.SortDirectionAscending {
			return iV < jV
		}

		return iV > jV
	})

	totalHits := len(productResults)

	pageAmount := int(0)

	if pageNumber < 0 {
		pageNumber = 0
	}
	start := (pageNumber - 1) * pageSize

	if start > len(productResults) {
		start = len(productResults)
	}
	stop := start + pageSize
	if stop > len(productResults) {
		stop = len(productResults)
	}
	productResults = productResults[start:stop]

	if pageSize > 0 {
		pageAmount = int(math.Ceil(float64(totalHits) / float64(pageSize)))
	}

	return &productDomain.SearchResult{
		Hits: productResults,
		Result: searchDomain.Result{
			SearchMeta: searchDomain.SearchMeta{
				NumResults: totalHits,
				NumPages:   pageAmount,
				Page:       pageNumber,
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

// addCategoryPath merges in the given categoryToAdd to the  passed currentExisting
// also adds new categories to the reverse index
func (r *InMemoryProductRepository) addCategoryPath(currentTreeNode *categoryDomain.TreeData, treeNodeToAdd *categoryDomain.TreeData) *categoryDomain.TreeData {
	if r.categoryTreeIndex == nil {
		r.categoryTreeIndex = make(map[string]*categoryDomain.TreeData)
	}
	// if its the first node then make as current node
	if currentTreeNode == nil {
		clone := *treeNodeToAdd
		currentTreeNode = &clone
		currentTreeNode.SubTreesData = nil
		r.categoryTreeIndex[currentTreeNode.CategoryCode] = currentTreeNode
	}
	if currentTreeNode.CategoryCode != treeNodeToAdd.CategoryCode {
		// No common root node - exit
		return currentTreeNode
	}
	for _, subTreeNodeToAdd := range treeNodeToAdd.SubTreesData {
		exists := false
		for _, existingSubTree := range currentTreeNode.SubTreesData {
			if existingSubTree.CategoryCode == subTreeNodeToAdd.CategoryCode {
				exists = true
				// match - proceed in recursion
				existingSubTree = r.addCategoryPath(existingSubTree, subTreeNodeToAdd)
			}
		}
		// subTreeNodeToAdd does not exist yet - so we merge it in:
		if !exists {
			currentTreeNode.SubTreesData = append(currentTreeNode.SubTreesData, subTreeNodeToAdd)
			r.updateCategoryIndex(subTreeNodeToAdd)
		}
	}
	return currentTreeNode
}

func (r *InMemoryProductRepository) updateCategoryIndex(tree *categoryDomain.TreeData) {
	r.categoryTreeIndex[tree.Code()] = tree
	for _, stree := range tree.SubTreesData {
		r.updateCategoryIndex(stree)
	}
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
		IsActive:     true,
		CategoryName: teaser.Name,
		CategoryCode: teaser.Code,
		SubTreesData: childs,
	}
	if teaser.Parent == nil {
		return currentCategory
	}

	return r.categoryTeaserToCategoryTree(*teaser.Parent, currentCategory)

}

func inSlice(list []string, search string) bool {
	for _, v := range list {
		if v == search {
			return true
		}
	}
	return false
}

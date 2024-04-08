package commercesearch

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/analysis/tokenizer/whitespace"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
)

type (

	// BleveRepository serves as a Repository of Products held in memory
	BleveRepository struct {
		index                            bleve.Index
		logger                           flamingo.Logger
		assignProductsToParentCategories bool
		cacheMutex                       sync.RWMutex
		cachedCategoryTree               categoryDomain.Tree
		cachedCategories                 map[string]categoryDomain.Category
		enableCategoryFacet              bool
		facetConfig                      []facetConfig
		sortConfig                       []sortConfig
	}

	facetConfig struct {
		AttributeCode string
		Amount        int
	}

	sortConfig struct {
		AttributeCode string
		AttributeType string
		Asc           bool
		Desc          bool
	}

	// bleveDocument envelop for indexed entities
	bleveDocument struct {
		Product  productDomain.BasicProduct
		Category *productDomain.CategoryTeaser
	}
)

const (
	attributeTypeNumeric         = "numeric"
	attributeTypeText            = "text"
	attributeTypeBool            = "bool"
	productType                  = "product"
	categoryType                 = "category"
	categoryIDPrefix             = "cat_"
	sourceFieldName              = "_source"
	typeFieldName                = "_type"
	fieldPrefixInIndexedDocument = "Product."
)

var (
	_ domain.ProductRepository  = &BleveRepository{}
	_ domain.CategoryRepository = &BleveRepository{}
	_ mapping.Classifier        = &bleveDocument{}
)

func init() {
	gob.Register(&productDomain.SimpleProduct{})
	gob.Register(&productDomain.ConfigurableProduct{})
}

func (b *bleveDocument) Type() string {
	if b.Category != nil {
		return categoryType
	}
	return productType
}

func (b *bleveDocument) getTypeField() *document.TextField {
	// Add type for phrase query
	return document.NewTextFieldCustom(
		typeFieldName, nil, []byte(b.Type()), document.IndexField|document.StoreField|document.IncludeTermVectors, nil)
}

// Inject dep
func (r *BleveRepository) Inject(logger flamingo.Logger, config *struct {
	AssignProductsToParentCategories bool         `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.productsToParentCategories,optional"`
	EnableCategoryFacet              bool         `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.enableCategoryFacet,optional"`
	FacetConfig                      config.Slice `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.facetConfig"`
	SortConfig                       config.Slice `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.sortConfig"`
}) *BleveRepository {
	r.logger = logger.WithField(flamingo.LogKeyModule, "flamingoCommerceAdapterStandalone.commercesearch").WithField(flamingo.LogKeyCategory, "bleve")
	if config != nil {
		r.assignProductsToParentCategories = config.AssignProductsToParentCategories
		r.enableCategoryFacet = config.EnableCategoryFacet
		var facetConfig []facetConfig
		err := config.FacetConfig.MapInto(&facetConfig)
		if err != nil {
			panic(err)
		}
		r.facetConfig = facetConfig

		var sortConfig []sortConfig
		err = config.SortConfig.MapInto(&sortConfig)
		if err != nil {
			panic(err)
		}
		r.sortConfig = sortConfig
	}
	return r
}

// PrepareIndex prepares bleve index with given configuration
func (r *BleveRepository) PrepareIndex(_ context.Context) error {

	// Init index
	mapping := bleve.NewIndexMapping()

	categoryCodeField := bleve.NewTextFieldMapping()
	categoryCodeField.Store = false
	categoryCodeField.IncludeTermVectors = false
	categoryCodeField.DocValues = false
	categoryCodeField.Store = false
	categoryCodeField.Index = false

	/* todo - enable persistent index ?
	indexName := "productRepIndex"
	if _, err := os.Stat(indexName); !os.IsNotExist(err) {
		r.Logger.Warn(indexName+" already exist!")
		bleve.Open(indexName)
	}*/
	// index, err := bleve.NewUsing("lily.bleve", mapping, scorch.Name, scorch.Name, nil)
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return err
	}
	r.index = index
	return nil
}

func (r *BleveRepository) getIndex() (bleve.Index, error) {
	if r.index == nil {
		return nil, errors.New("index not prepared")
	}
	return r.index, nil

}

// DocumentsCount returns the number of documents in the index
func (r *BleveRepository) DocumentsCount() int64 {
	c, _ := r.index.DocCount()

	return int64(c)
}

// UpdateByCategoryTeasers updates or appends a category to the Product Repository
func (r *BleveRepository) UpdateByCategoryTeasers(_ context.Context, categoryTeasers []productDomain.CategoryTeaser) error {
	index, err := r.getIndex()
	if err != nil {
		return err
	}
	batch := index.NewBatch()
	for _, categoryTeaser := range categoryTeasers {
		bleveCatDocuments, err := r.categoryTeaserToBleve(categoryTeaser, nil)
		if err != nil {
			return err
		}
		for _, bleveCatDocument := range bleveCatDocuments {
			err = batch.IndexAdvanced(bleveCatDocument)
			if err != nil {
				return err
			}
		}

	}

	return index.Batch(batch)

}

// ClearCategories clears given ids from repo
func (r *BleveRepository) ClearCategories(_ context.Context, categoryIds []string) error {
	return nil
}

// ClearProducts from Product Repository
func (r *BleveRepository) ClearProducts(_ context.Context, products []string) error {
	return nil
}

// UpdateProducts products to the Product Repository
func (r *BleveRepository) UpdateProducts(_ context.Context, products []productDomain.BasicProduct) error {
	index, err := r.getIndex()
	if err != nil {
		return err
	}

	for _, product := range products {
		// to receive original
		if product.BaseData().MarketPlaceCode == "" {
			return fmt.Errorf("No marketplace code %v, %v", product.GetIdentifier(), product.BaseData().Title)
		}

		bleveDocuments, err := r.productToBleveDocs(product)

		if err != nil {
			return err
		}
		// index to bleve
		batch := index.NewBatch()
		for _, bleveDocument := range bleveDocuments {
			err = batch.IndexAdvanced(bleveDocument)
			if err != nil {
				return err
			}
		}

		err = index.Batch(batch)
		if err != nil {
			return err
		}
	}
	return nil
}

// productToBleveDocs returns the Product and Category documents to be indexed
func (r *BleveRepository) productToBleveDocs(product productDomain.BasicProduct) ([]*document.Document, error) {
	var bleveDocuments []*document.Document
	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}

	productEncoded, err := r.encodeProduct(product)
	if err != nil {
		return nil, err
	}
	indexDocument := bleveDocument{Product: product}
	bleveProductDocument := document.NewDocument(product.BaseData().MarketPlaceCode)
	err = index.Mapping().MapDocument(bleveProductDocument, indexDocument)
	if err != nil {
		return nil, err
	}

	// Add _source field with Gob encoded content (to restore original)
	field := document.NewTextFieldWithIndexingOptions(
		sourceFieldName, nil, productEncoded, document.StoreField)
	bleveProductDocument = bleveProductDocument.AddField(field)

	for _, sort := range r.sortConfig {
		var field document.Field
		switch sort.AttributeType {
		case attributeTypeNumeric:
			val, _ := strconv.ParseFloat(product.BaseData().Attribute(sort.AttributeCode).Value(), 64)
			field = document.NewNumericField(
				fieldPrefixInIndexedDocument+"sort."+sort.AttributeCode, nil, val)
		case attributeTypeText:
			field = document.NewTextFieldCustom(
				fieldPrefixInIndexedDocument+"sort."+sort.AttributeCode, nil, []byte(product.BaseData().Attribute(sort.AttributeCode).Value()), document.IndexField, nil)
		case attributeTypeBool:
			val, _ := strconv.ParseBool(product.BaseData().Attribute(sort.AttributeCode).Value())
			field = document.NewBooleanField(
				fieldPrefixInIndexedDocument+"sort."+sort.AttributeCode, nil, val)
		}
		bleveProductDocument = bleveProductDocument.AddField(field)
	}

	// Add price Field to support sorting by price
	priceField := document.NewNumericField(
		fieldPrefixInIndexedDocument+"sort.price", nil, product.TeaserData().TeaserPrice.GetFinalPrice().FloatAmount())
	bleveProductDocument = bleveProductDocument.AddField(priceField)

	//  Add category field for category facet and filter
	tok, err := whitespace.TokenizerConstructor(nil, nil)
	if err != nil {
		return nil, err
	}
	analyser := &analysis.Analyzer{
		CharFilters:  nil,
		Tokenizer:    tok,
		TokenFilters: nil,
	}

	allCategories, allCategoryPaths := r.categoryParentCodes(product.BaseData().MainCategory)
	for _, c := range product.BaseData().Categories {
		codes, paths := r.categoryParentCodes(c)
		allCategories = append(allCategories, codes...)
		allCategoryPaths = append(allCategoryPaths, paths...)
	}
	// Add "Product.Facet.Categorycode" - Used for CategoryFilter
	categoryField := document.NewTextFieldCustom(
		fieldPrefixInIndexedDocument+"Facet.Categorycode", nil, []byte(strings.Join(allCategories, " ")), document.IndexField|document.StoreField|document.IncludeTermVectors, analyser)
	bleveProductDocument = bleveProductDocument.AddField(categoryField)

	// Add "Product.Facet.CategoryPaths" - Used for CategoryFilter
	categoryPathField := document.NewTextFieldCustom(
		fieldPrefixInIndexedDocument+"Facet.CategoryPaths", nil, []byte(strings.Join(allCategoryPaths, " ")), document.IndexField|document.StoreField|document.IncludeTermVectors, analyser)
	bleveProductDocument = bleveProductDocument.AddField(categoryPathField)

	// Add Configured Facet Attributes
	for _, facetConfig := range r.facetConfig {
		attributeField := document.NewTextFieldCustom(
			fieldPrefixInIndexedDocument+"Facet.Attribute."+facetConfig.AttributeCode, nil, []byte(product.BaseData().Attribute(facetConfig.AttributeCode).Label), document.IndexField|document.StoreField|document.IncludeTermVectors, analyser)
		bleveProductDocument = bleveProductDocument.AddField(attributeField)
	}

	// Add Type Field
	bleveProductDocument = bleveProductDocument.AddField(indexDocument.getTypeField())

	for _, va := range bleveProductDocument.Fields {
		_ = va
		// fmt.Printf("\n bleveDocument Fields: %#v : %v / tv: %v",va.Name(),string(va.Value()),va.Options().String())
	}
	bleveDocuments = append(bleveDocuments, bleveProductDocument)
	return bleveDocuments, nil
}

func (r *BleveRepository) categoryParentCodes(teaser productDomain.CategoryTeaser) (categoryCodes []string, categoryPaths []string) {
	codes := []string{teaser.Code}
	paths := []string{teaser.CPath()}
	if !r.assignProductsToParentCategories {
		return codes, paths
	}
	if teaser.Parent != nil {
		parentCodes, parentPaths := r.categoryParentCodes(*teaser.Parent)
		codes = append(codes, parentCodes...)
		paths = append(paths, parentPaths...)
	}
	return codes, paths
}

// categoryTeaserToBleve returns bleve documents for type category for the given TeaserData (called recursive with Parent)
func (r *BleveRepository) categoryTeaserToBleve(categoryTeaser productDomain.CategoryTeaser, alreadyAddedBleveDocs []*document.Document) ([]*document.Document, error) {
	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}
	indexDocument := bleveDocument{Category: &categoryTeaser}

	bleveCatDocument := document.NewDocument(categoryIDPrefix + categoryTeaser.Code)
	err = index.Mapping().MapDocument(bleveCatDocument, indexDocument)
	if err != nil {
		return nil, err
	}
	bleveCatDocument = bleveCatDocument.AddField(indexDocument.getTypeField())
	if categoryTeaser.Parent == nil || categoryTeaser.Parent.Code == categoryTeaser.Code {
		isRootField := document.NewBooleanFieldWithIndexingOptions(
			"Category.IsRoot", nil, true, document.IndexField|document.StoreField|document.IncludeTermVectors)
		bleveCatDocument = bleveCatDocument.AddField(isRootField)
	}

	if categoryTeaser.Parent != nil {
		parentCatField := document.NewTextFieldCustom(
			"Category.Parent.Code", nil, []byte(categoryTeaser.Parent.Code), document.IndexField|document.StoreField|document.IncludeTermVectors, nil)
		bleveCatDocument = bleveCatDocument.AddField(parentCatField)
	}

	/*
		for _,va := range bleveCatDocument.Fields {
			_ = va
			fmt.Printf("\n bleveCatDocument Fields: %#v : %v / tv: %v",va.Name(),string(va.Value()),va.Options().String())
		}
	*/

	alreadyAddedBleveDocs = append(alreadyAddedBleveDocs, bleveCatDocument)
	if categoryTeaser.Parent != nil {
		return r.categoryTeaserToBleve(*categoryTeaser.Parent, alreadyAddedBleveDocs)
	}
	return alreadyAddedBleveDocs, nil
}

func (r *BleveRepository) bleveHitToProduct(hit *search.DocumentMatch) (productDomain.BasicProduct, error) {
	b, ok := hit.Fields[sourceFieldName]

	if !ok {
		return nil, errors.New("_source field missing in hit")
	}
	return r.decodeProduct([]byte(fmt.Sprintf("%v", b)))
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *BleveRepository) FindByMarketplaceCode(_ context.Context, marketplaceCode string) (productDomain.BasicProduct, error) {
	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}

	docIDQuery := query.NewDocIDQuery([]string{marketplaceCode})
	searchRequest := bleve.NewSearchRequest(docIDQuery)
	searchRequest.Fields = append(searchRequest.Fields, sourceFieldName)
	searchResult, err := index.Search(searchRequest)

	if err != nil {
		return nil, err
	}
	if searchResult.Total < 1 {
		return nil, productDomain.ProductNotFound{MarketplaceCode: marketplaceCode}
	}

	return r.bleveHitToProduct(searchResult.Hits[0])
}

// CategoryTree returns tree
func (r *BleveRepository) CategoryTree(_ context.Context, code string) (categoryDomain.Tree, error) {

	if code == "" && r.cachedCategoryTree != nil {
		return r.cachedCategoryTree, nil
	}
	category, err := r.Category(nil, code)
	if err != nil {
		return nil, err
	}

	rootTreeNode := mapCatToTree(category)
	subTrees, err := r.subTrees(rootTreeNode)
	if err != nil {
		return nil, err
	}
	rootTreeNode.SubTreesData = subTrees
	if code == "" {
		r.cachedCategoryTree = rootTreeNode
	}

	return rootTreeNode, nil
}

func mapCatToTree(category categoryDomain.Category) *categoryDomain.TreeData {
	return &categoryDomain.TreeData{
		CategoryCode:          category.Code(),
		CategoryName:          category.Name(),
		CategoryPath:          category.Path(),
		CategoryDocumentCount: 0,
		SubTreesData:          nil,
		IsActive:              false,
	}
}

func (r *BleveRepository) subTrees(parentNode *categoryDomain.TreeData) ([]*categoryDomain.TreeData, error) {
	// fmt.Printf("\n subtrees for %v",parentNode.CategoryCode)
	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}

	squery := bleve.NewConjunctionQuery(bleve.NewPhraseQuery([]string{categoryType}, typeFieldName), bleve.NewPhraseQuery([]string{parentNode.CategoryCode}, "Category.Parent.Code"))
	searchRequest := bleve.NewSearchRequestOptions(squery, 100, 0, false)
	searchRequest.Fields = append(searchRequest.Fields, sourceFieldName, "Category.Code", "Category.Name", "Category.Parent.Code", "Category.Path")

	searchResults, err := index.Search(searchRequest)

	if err != nil {
		return nil, err
	}

	var subTreeNodes []*categoryDomain.TreeData
	for _, hit := range searchResults.Hits {
		treeNode := mapCatToTree(mapHitToCategory(hit))
		subTreeNodesOfSub, err := r.subTrees(treeNode)
		if err != nil {
			return nil, err
		}
		treeNode.SubTreesData = subTreeNodesOfSub
		subTreeNodes = append(subTreeNodes, treeNode)
	}
	return subTreeNodes, nil
}

// Category receives indexed categories
func (r *BleveRepository) Category(_ context.Context, code string) (categoryDomain.Category, error) {

	r.cacheMutex.RLock()
	if cat, ok := r.cachedCategories[code]; ok {
		r.cacheMutex.RUnlock()
		return cat, nil
	}
	r.cacheMutex.RUnlock()

	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}

	var squery query.Query
	if code != "" {
		squery = query.NewDocIDQuery([]string{categoryIDPrefix + code})
	} else {
		boolQ := bleve.NewBoolFieldQuery(true)
		boolQ.SetField("Category.IsRoot")
		squery = bleve.NewConjunctionQuery(bleve.NewPhraseQuery([]string{categoryType}, typeFieldName), boolQ)
	}

	searchRequest := bleve.NewSearchRequestOptions(squery, 1, 0, false)
	searchRequest.Fields = append(searchRequest.Fields, sourceFieldName, "Category.Code", "Category.Name", "Category.Parent.Code", "Category.Path")

	searchResults, err := index.Search(searchRequest)

	if err != nil {
		return nil, err
	}

	if searchResults.Total != 1 {
		return nil, categoryDomain.ErrNotFound
	}
	cat := mapHitToCategory(searchResults.Hits[0])
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	if r.cachedCategories == nil {
		r.cachedCategories = make(map[string]categoryDomain.Category)
	}
	r.cachedCategories[code] = cat
	return mapHitToCategory(searchResults.Hits[0]), nil

}

func mapHitToCategory(hit *search.DocumentMatch) categoryDomain.Category {
	return &categoryDomain.CategoryData{
		CategoryCode:       fmt.Sprintf("%v", hit.Fields["Category.Code"]),
		CategoryName:       fmt.Sprintf("%v", hit.Fields["Category.Name"]),
		CategoryPath:       fmt.Sprintf("%v", hit.Fields["Category.Path"]),
		IsPromoted:         false,
		IsActive:           false,
		CategoryMedia:      nil,
		CategoryTypeCode:   "",
		CategoryAttributes: nil,
		Promotion:          categoryDomain.Promotion{},
	}
}

// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *BleveRepository) Find(_ context.Context, filters ...searchDomain.Filter) (*productDomain.SearchResult, error) {

	index, err := r.getIndex()
	if err != nil {
		return nil, err
	}

	var mainQuery query.Query

	currentPage := 1
	pageSize := 100
	sortingField := ""
	sortingDesc := true

	// First check if we have a human query filter:
	for _, filter := range filters {
		switch f := filter.(type) {
		case *searchDomain.QueryFilter:
			mainQuery = bleve.NewQueryStringQuery(f.Query())
		}
	}

	var filterQueryParts []query.Query

	for _, filter := range filters {
		r.logger.Info("Find ", fmt.Sprintf("%T %#v", filter, filter))
		switch f := filter.(type) {
		case *searchDomain.KeyValueFilter:
			if f.Key() == "category" {
				filterQueryParts = append(filterQueryParts, newDisjunctionTermQuery(f.KeyValues(), fieldPrefixInIndexedDocument+"Facet.Categorycode"))
			} else {
				filterQueryParts = append(filterQueryParts, newDisjunctionTermQuery(f.KeyValues(), fieldPrefixInIndexedDocument+"Facet.Attribute."+f.Key()))
			}
		case categoryDomain.CategoryFacet:
			filterQueryParts = append(filterQueryParts, newDisjunctionTermQuery([]string{f.CategoryCode}, fieldPrefixInIndexedDocument+"Facet.Categorycode"))
		case *categoryDomain.CategoryFacet:
			filterQueryParts = append(filterQueryParts, newDisjunctionTermQuery([]string{f.CategoryCode}, fieldPrefixInIndexedDocument+"Facet.Categorycode"))
		case *searchDomain.PaginationPage:
			currentPage = f.GetPage()
		case *searchDomain.PaginationPageSize:
			pageSize = f.GetPageSize()
		case *searchDomain.SortFilter:
			sortingField = fieldPrefixInIndexedDocument + "sort." + f.Field()
			sortingDesc = f.Descending()
		}
	}

	if mainQuery == nil && len(filterQueryParts) > 0 {
		mainQuery = bleve.NewMatchAllQuery()
	}
	if mainQuery != nil {
		filterQueryParts = append(filterQueryParts, mainQuery)
	}
	filterQueryParts = append(filterQueryParts, bleve.NewPhraseQuery([]string{productType}, typeFieldName))
	facetsRequests := make(bleve.FacetsRequest)
	if r.enableCategoryFacet {
		facetRequest := bleve.NewFacetRequest(fieldPrefixInIndexedDocument+"Facet.CategoryPaths", 100)
		facetsRequests["category"] = facetRequest
	}
	for _, facetConfig := range r.facetConfig {
		facetRequest := bleve.NewFacetRequest(fieldPrefixInIndexedDocument+"Facet.Attribute."+facetConfig.AttributeCode, facetConfig.Amount)
		facetsRequests[facetConfig.AttributeCode] = facetRequest
	}
	conjunctionQuery := bleve.NewConjunctionQuery(filterQueryParts...)
	from := (currentPage - 1) * pageSize
	searchRequest := bleve.NewSearchRequestOptions(conjunctionQuery, pageSize, from, false)
	searchRequest.Facets = facetsRequests
	searchRequest.Fields = append(searchRequest.Fields, sourceFieldName)
	if sortingField != "" {
		searchRequest.SortByCustom(search.SortOrder{
			&search.SortField{
				Field:   sortingField,
				Missing: search.SortFieldMissingLast,
				Desc:    sortingDesc,
			},
		})
	}

	searchResults, err := index.Search(searchRequest)

	if err != nil {
		return nil, err
	}

	result := r.mapBleveResultToResult(searchResults)
	markActiveFacets(filters, result)

	return result, nil
}

// newDisjunctionTermQuery creates a disjunctive term query, meaning that any of the provided terms can match
func newDisjunctionTermQuery(terms []string, field string) *query.DisjunctionQuery {
	termQuery := bleve.NewDisjunctionQuery()
	for _, term := range terms {
		termQuery.AddQuery(bleve.NewPhraseQuery([]string{term}, field))
	}
	return termQuery
}

func markActiveFacets(filters []searchDomain.Filter, result *productDomain.SearchResult) {
	for _, filter := range filters {
		if f, ok := filter.(*searchDomain.KeyValueFilter); ok {
			for i, facetItem := range result.Facets[f.Key()].Items {
				for _, selectedValue := range f.KeyValues() {
					if facetItem.Value == selectedValue {
						facetItem.Selected = true
						facetItem.Active = true
					}
				}
				result.Facets[f.Key()].Items[i] = facetItem
			}
		}
	}
}

func (r *BleveRepository) mapBleveResultToResult(searchResults *bleve.SearchResult) *productDomain.SearchResult {
	pageAmount := 0
	pageSize := searchResults.Request.Size
	currentPage := 1

	if pageSize > 0 {
		pageAmount = int(math.Ceil(float64(searchResults.Total) / float64(pageSize)))
		if searchResults.Request.From > 0 {
			currentPage = 1 + searchResults.Request.From/pageSize
		}
	}
	var productResults []productDomain.BasicProduct
	resultFacetCollection := make(searchDomain.FacetCollection)
	facetResultForConfiguredName := func(name string) *search.FacetResult {
		for k, f := range searchResults.Facets {
			if name == k {
				return f
			}
		}
		return nil
	}
	if r.enableCategoryFacet {
		facetResult := facetResultForConfiguredName("category")
		if facetResult == nil {
			r.logger.Warn("No facet result for category facet ")
		}
		var constructedItems []*searchDomain.FacetItem

		for _, termFacetTerms := range facetResult.Terms {
			pathSegments := strings.Split(termFacetTerms.Term, "/")
			if len(pathSegments) < 2 {
				continue
			}
			pathSegments = pathSegments[1:]
			constructedItems = r.constructCategoryTreeFacet(constructedItems, pathSegments, int64(termFacetTerms.Count))
		}

		facet := searchDomain.Facet{
			Type:     searchDomain.TreeFacet,
			Name:     "category",
			Label:    "category",
			Items:    constructedItems,
			Position: 0,
		}
		resultFacetCollection["category"] = facet
	}

	for _, facetConfig := range r.facetConfig {
		facetResult := facetResultForConfiguredName(facetConfig.AttributeCode)
		if facetResult == nil {
			r.logger.Warn("No facet result for configured facet ", facetConfig.AttributeCode)
		}
		facet := searchDomain.Facet{
			Type:     searchDomain.ListFacet,
			Name:     facetConfig.AttributeCode,
			Label:    facetConfig.AttributeCode,
			Position: 0,
		}
		for _, termFacetTerms := range facetResult.Terms {
			facet.Items = append(facet.Items, &searchDomain.FacetItem{
				Label:    termFacetTerms.Term,
				Value:    termFacetTerms.Term,
				Active:   false,
				Selected: false,
				Count:    int64(termFacetTerms.Count),
			})
		}
		resultFacetCollection[facetConfig.AttributeCode] = facet
	}
	for _, hit := range searchResults.Hits {
		product, err := r.bleveHitToProduct(hit)
		if err != nil {
			r.logger.Error(err)
			continue
		}
		productResults = append(productResults, product)
	}

	sortOptions := []searchDomain.SortOption{
		{
			Label:        "Price",
			Field:        "price",
			SelectedAsc:  false,
			SelectedDesc: false,
			Asc:          "price",
			Desc:         "price",
		},
	}

	for _, s := range r.sortConfig {
		sortOptions = append(sortOptions, searchDomain.SortOption{
			Label: s.AttributeCode,
			Field: s.AttributeCode,
			Asc: func() string {
				if s.Asc {
					return s.AttributeCode
				}
				return ""
			}(),
			Desc: func() string {
				if s.Desc {
					return s.AttributeCode
				}
				return ""
			}(),
		})
	}

	return &productDomain.SearchResult{
		Hits: productResults,
		Result: searchDomain.Result{
			Facets: resultFacetCollection,
			SearchMeta: searchDomain.SearchMeta{
				NumResults:  int(searchResults.Total),
				NumPages:    pageAmount,
				Page:        currentPage,
				SortOptions: sortOptions,
			}},
	}
}

func (r *BleveRepository) encodeProduct(product productDomain.BasicProduct) ([]byte, error) {
	var mess bytes.Buffer
	if sp, ok := product.(productDomain.SimpleProduct); ok {
		product = &sp
	}
	if cp, ok := product.(productDomain.ConfigurableProduct); ok {
		product = &cp
	}

	enc := gob.NewEncoder(&mess)
	err := enc.Encode(&product)
	if err != nil {
		return nil, err
	}
	return mess.Bytes(), nil
}

func (r *BleveRepository) constructCategoryTreeFacet(parentSlice []*searchDomain.FacetItem, remainingPathSegments []string, count int64) []*searchDomain.FacetItem {
	currentSegment := remainingPathSegments[0]
	isLast := len(remainingPathSegments) == 1

	var foundItem *searchDomain.FacetItem
	for _, item := range parentSlice {
		if item.Value == currentSegment {
			foundItem = item
		}
	}
	if foundItem == nil {
		label := currentSegment
		cat, err := r.Category(nil, currentSegment)
		if err != nil {
			r.logger.Warn("Cannot build Tree Facet du to missing category details", err)
		} else {
			label = cat.Name()
		}
		foundItem = &searchDomain.FacetItem{
			Label: label,
			Value: currentSegment,
		}
		parentSlice = append(parentSlice, foundItem)
	}
	if isLast {
		// use count if path matches
		foundItem.Count = count
		return parentSlice
	}
	foundItem.Items = r.constructCategoryTreeFacet(foundItem.Items, remainingPathSegments[1:], count)
	return parentSlice
}

func (r *BleveRepository) decodeProduct(b []byte) (productDomain.BasicProduct, error) {
	buffer := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buffer)

	var sp productDomain.BasicProduct
	err := dec.Decode(&sp)
	if err != nil {
		return nil, err
	}
	if p, ok := sp.(*productDomain.SimpleProduct); ok {
		sp = *p
	}
	if p, ok := sp.(*productDomain.ConfigurableProduct); ok {
		sp = *p
	}
	return sp, nil
}

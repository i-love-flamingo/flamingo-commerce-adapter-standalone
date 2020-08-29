package productsearch

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/analysis/tokenizer/whitespace"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"math"
	"strings"
)

type (

	// BleveRepository serves as a Repository of Products held in memory
	BleveRepository struct {
		index                            bleve.Index
		logger                           flamingo.Logger
		assignProductsToParentCategories bool
		cachedCategoryTree categoryDomain.Tree
	}

	//bleveDocument - envelop for indexed entities
	bleveDocument struct {
		Product  productDomain.BasicProduct
		Category *productDomain.CategoryTeaser
	}
)

const productType = "product"
const categoryType = "category"
const categoryIDPrefix = "cat_"

const sourceFieldName = "_source"

const typeFieldName = "_type"

const fieldPrefixInIndexedDocument = "Product."

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
	//Add type for phrase query
	return document.NewTextFieldCustom(
		typeFieldName, nil, []byte(b.Type()), document.IndexField|document.StoreField|document.IncludeTermVectors, nil)
}

//Inject dep
func (r *BleveRepository) Inject(logger flamingo.Logger, config *struct {
	AssignProductsToParentCategories bool `inject:config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.productsToParentCategories,optional`
}) *BleveRepository {
	r.logger = logger
	if config != nil {
		r.assignProductsToParentCategories = config.AssignProductsToParentCategories
	}
	return r
}

//PrepareIndex - prepares bleve index with given configuration
func (r *BleveRepository) PrepareIndex(_ context.Context) error {

	//Init index
	mapping := bleve.NewIndexMapping()

	categoryCodeField := bleve.NewTextFieldMapping()
	categoryCodeField.Store = false
	categoryCodeField.IncludeTermVectors = false
	categoryCodeField.DocValues = false
	categoryCodeField.Store = false
	categoryCodeField.Index = false
	/*
		productMapping := bleve.NewDocumentMapping()
		productMapping.AddFieldMappingsAt(fieldPrefixInIndexedDocument+"MMMainCategory.Code",categoryCodeField)
		mapping.AddDocumentMapping(productType, productMapping)
		mapping.DefaultMapping.AddFieldMappingsAt(fieldPrefixInIndexedDocument+"MMMainCategory.Code",categoryCodeField)
		mapping.AddDocumentMapping(productType, productMapping)
	*/

	/* todo - enable persistent index ?
	indexName := "productRepIndex"
	if _, err := os.Stat(indexName); !os.IsNotExist(err) {
		r.Logger.Warn(indexName+" already exist!")
		bleve.Open(indexName)
	}*/
	//index, err := bleve.NewUsing("lily.bleve", mapping, scorch.Name, scorch.Name, nil)
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return err
	}
	r.index = index
	return nil
}

func (r *BleveRepository) getIndex() (bleve.Index, error) {
	if r.index == nil {
		return nil, errors.New("Index not prepared")
	}
	return r.index, nil

}

// UpdateByCategoryTeasers - updates or appends a category to the Product Repository
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

// UpdateProducts  products to the Product Repository
func (r *BleveRepository) UpdateProducts(_ context.Context, products []productDomain.BasicProduct) error {
	index, err := r.getIndex()
	if err != nil {
		return err
	}

	for _, product := range products {
		//to receive original
		if product.BaseData().MarketPlaceCode == "" {
			return errors.New("No marketplace code ")
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

//productToBleveDocs - returns the Product and Category documents to be indexed
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

	//Add _source field with Gib ncoded content (to restore original)
	field := document.NewTextFieldWithIndexingOptions(
		sourceFieldName, nil, productEncoded, document.StoreField)
	bleveProductDocument = bleveProductDocument.AddField(field)

	//Add price Field to support sorting by price
	priceField := document.NewNumericField(
		fieldPrefixInIndexedDocument+"Sort.Price", nil, product.TeaserData().TeaserPrice.GetFinalPrice().FloatAmount())
	bleveProductDocument = bleveProductDocument.AddField(priceField)

	//Add title Field to support sorting by raw title (without analysers)
	titleSortField := document.NewTextFieldCustom(
		fieldPrefixInIndexedDocument+"Sort.Title", nil, []byte(product.BaseData().Title),
		document.IndexField, nil)
	bleveProductDocument = bleveProductDocument.AddField(titleSortField)

	//Add category field for category facet
	tok, err := whitespace.TokenizerConstructor(nil, nil)
	if err != nil {
		return nil, err
	}
	analyser := &analysis.Analyzer{
		CharFilters:  nil,
		Tokenizer:    tok,
		TokenFilters: nil,
	}

	allCategories := r.categoryParentCodes(product.BaseData().MainCategory)
	for _, c := range product.BaseData().Categories {
		allCategories = append(allCategories, r.categoryParentCodes(c)...)
	}
	categoryField := document.NewTextFieldCustom(
		fieldPrefixInIndexedDocument+"Facet.Categorycode", nil, []byte(strings.Join(allCategories, " ")), document.IndexField|document.StoreField|document.IncludeTermVectors, analyser)
	bleveProductDocument = bleveProductDocument.AddField(categoryField)

	bleveProductDocument = bleveProductDocument.AddField(indexDocument.getTypeField())

	for _, va := range bleveProductDocument.Fields {
		_ = va
		//fmt.Printf("\n bleveDocument Fields: %#v : %v / tv: %v",va.Name(),string(va.Value()),va.Options().String())
	}
	bleveDocuments = append(bleveDocuments, bleveProductDocument)
	return bleveDocuments, nil
}

func (r *BleveRepository) categoryParentCodes(teaser productDomain.CategoryTeaser) []string {
	codes := []string{teaser.Code}
	if !r.assignProductsToParentCategories {
		return codes
	}
	if teaser.Parent != nil {
		parentCode := r.categoryParentCodes(*teaser.Parent)
		codes = append(codes, parentCode...)
	}
	return codes
}

//categoryTeaserToBleve returns bleve documents for type category for the given Teaserdata (called recursive with Parent)
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

//CategoryTree returns tree
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

//mapCatToTree
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
	//fmt.Printf("\n subtrees for %v",parentNode.CategoryCode)
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

//Category - receives indexed categories
func (r *BleveRepository) Category(_ context.Context, code string) (categoryDomain.Category, error) {

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

	return mapHitToCategory(searchResults.Hits[0]), nil

}

//mapHitToCategory
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
	var productResults []productDomain.BasicProduct
	currentPage := int(1)
	pageSize := int(100)
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
		r.logger.Warn(fmt.Sprintf("\n Filter PPPPPPP %#v %T", filter, filter))
		switch f := filter.(type) {
		case *searchDomain.KeyValueFilter:
			filterQueryParts = append(filterQueryParts, bleve.NewPhraseQuery(f.KeyValues(), f.Key()))
		case categoryDomain.CategoryFacet:
			term := bleve.NewPhraseQuery([]string{f.CategoryCode}, fieldPrefixInIndexedDocument+"Facet.Categorycode")
			filterQueryParts = append(filterQueryParts, term)
		case *categoryDomain.CategoryFacet:
			term := bleve.NewPhraseQuery([]string{f.CategoryCode}, fieldPrefixInIndexedDocument+"Facet.Categorycode")
			filterQueryParts = append(filterQueryParts, term)
		case *searchDomain.PaginationPage:
			currentPage = f.GetPage()
		case *searchDomain.PaginationPageSize:
			pageSize = f.GetPageSize()
		case *searchDomain.SortFilter:
			sortingField = fieldPrefixInIndexedDocument + "Sort." + f.Field()
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
	query := bleve.NewConjunctionQuery(filterQueryParts...)

	searchRequest := bleve.NewSearchRequestOptions(query, pageSize, currentPage-1, false)
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
	for _, hit := range searchResults.Hits {
		product, err := r.bleveHitToProduct(hit)
		if err != nil {
			r.logger.Error(err)
			continue
		}
		productResults = append(productResults, product)
	}

	pageAmount := 0
	if pageSize > 0 {
		pageAmount = int(math.Ceil(float64(searchResults.Total) / float64(pageSize)))
	}
	return &productDomain.SearchResult{
		Hits: productResults,
		Result: searchDomain.Result{
			SearchMeta: searchDomain.SearchMeta{
				NumResults: int(searchResults.Total),
				NumPages:   pageAmount,
				Page:       currentPage,
			}},
	}, nil
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

func (r *BleveRepository) decodeProduct(b []byte) (productDomain.BasicProduct, error) {
	messout := bytes.NewBuffer(b)
	dec := gob.NewDecoder(messout)

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

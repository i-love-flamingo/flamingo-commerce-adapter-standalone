package productSearch

import (
	"errors"
	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/blevesearch/bleve/search/query"
	"math"
	"sync"

	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"github.com/blevesearch/bleve"
)

type (

	// BleveRepository serves as a Repository of Products held in memory
	BleveRepository struct {
		index bleve.Index
		//marketplaceCodeIndex - index to get products from marketplaceCode
		marketplaceCodeIndex map[string]productDomain.BasicProduct
		addReadMutex                     sync.RWMutex
		Logger flamingo.Logger `inject:""`
	}

)


func (r *BleveRepository) PrepareIndex() error {
	if r.marketplaceCodeIndex == nil {
		r.marketplaceCodeIndex = make(map[string]productDomain.BasicProduct)
	}
	mapping := bleve.NewIndexMapping()
	var err error
	/* todo - enable persistent index ?
	indexName := "productRepIndex"
	if _, err := os.Stat(indexName); !os.IsNotExist(err) {
		r.Logger.Warn(indexName+" already exist!")
		bleve.Open(indexName)
	}*/
	r.index, err = bleve.NewMemOnly(mapping)
	return err
}

// Add appends a product to the Product Repository
func (r *BleveRepository) Add(product productDomain.BasicProduct) error {
	r.addReadMutex.Lock()
	defer r.addReadMutex.Unlock()
	//to receive original
	if (product.BaseData().MarketPlaceCode == "" ) {
		return errors.New("No marketplace code ")
	}
	if r.marketplaceCodeIndex[product.BaseData().MarketPlaceCode] != nil {
		r.Logger.Warn("Duplicate for marketplace code " + product.BaseData().MarketPlaceCode)
	}
	r.marketplaceCodeIndex[product.BaseData().MarketPlaceCode] = product
	// index to bleve
	return r.index.Index(product.BaseData().MarketPlaceCode, product)
}

// FindByMarketplaceCode returns a product struct for the given marketplaceCode
func (r *BleveRepository) FindByMarketplaceCode(marketplaceCode string) (productDomain.BasicProduct, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()
	if product, ok := r.marketplaceCodeIndex[marketplaceCode]; ok {
		return product, nil
	}
	return nil, productDomain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

func (r *BleveRepository) CategoryTree(code string) (categoryDomain.Tree, error)  {
	return nil, nil
}


func (r *BleveRepository) Category(code string) (categoryDomain.Category, error)  {

	return nil, errors.New("not found")
}


// Find returns a slice of product structs filtered from the product repository after applying the given filters
func (r *BleveRepository) Find(filters ...searchDomain.Filter) (*productDomain.SearchResult, error) {
	r.addReadMutex.RLock()
	defer r.addReadMutex.RUnlock()




	var queryParts []query.Query
	var productResults []productDomain.BasicProduct
	pageDefault := int(1)
	pageSizeDefault := int(100)
	for _, filter := range filters {
		switch f := filter.(type) {
		case *searchDomain.KeyValueFilter:
			queryParts = append(queryParts, bleve.NewPhraseQuery(f.KeyValues(),f.Key()))
		case *searchDomain.QueryFilter:
			queryParts = append(queryParts, bleve.NewQueryStringQuery(f.Query()))
		case categoryDomain.CategoryFacet:
			queryParts = append(queryParts, bleve.NewPhraseQuery([]string{f.CategoryCode},"MainCategory.Code"))
		case *searchDomain.PaginationPage:
			pageDefault = f.GetPage()
		case *searchDomain.PaginationPageSize:
			pageSizeDefault = f.GetPageSize()

		}
	}

	query := bleve.NewConjunctionQuery(queryParts...)
	searchRequest := bleve.NewSearchRequestOptions(query, pageSizeDefault, pageDefault-1, false)
	// todo support searchRequest.Sort
	searchResults, err := r.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	for _,hit := range searchResults.Hits {
		product , err := r.FindByMarketplaceCode(hit.ID)
		if err != nil {
			continue
		}
		productResults = append(productResults,product)
	}

	pageAmount := 0
	if pageSizeDefault > 0 {
		pageAmount = int(math.Ceil(float64(searchResults.Size()) / float64(pageSizeDefault)))
	}
	return &productDomain.SearchResult{
		Hits: productResults,
		Result: searchDomain.Result{
			SearchMeta: searchDomain.SearchMeta{
				NumResults: searchResults.Size(),
				NumPages:   pageAmount,
			}},
	}, nil
}

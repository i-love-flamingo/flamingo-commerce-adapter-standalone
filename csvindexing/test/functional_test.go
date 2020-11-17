package test_test

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"sort"
	"testing"

	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/require"

	domain2 "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	csvcommerceLoader "flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/infrastructure/commercesearch"

	"flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"github.com/stretchr/testify/assert"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/productsearch"
)

func TestFactoryCanBuildSimpleTest(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t, "products2.csv")

	rootCat, err := rep.Category(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "master", rootCat.Code())

	rootNode, err := rep.CategoryTree(context.Background(), "")
	assert.NoError(t, err)
	assert.Equal(t, "master", rootNode.Code())
	product, err := rep.FindByMarketplaceCode(context.Background(), "1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "1000000", product.BaseData().MarketPlaceCode)
	assert.Equal(t, "Hello Kitty Candy Cup", product.BaseData().Title)
	assert.Equal(t, "accessories", product.BaseData().MainCategory.Code)
}

func TestFactoryBugForMissingCat(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t, "products_en.csv")

	rootCat, err := rep.Category(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "master", rootCat.Code())

	tabletsC, err := rep.Category(context.Background(), "tablets")
	require.NoError(t, err)
	assert.Equal(t, "tablets", tabletsC.Code())

	pclaptopsCat, err := rep.Category(context.Background(), "pc_laptops")
	require.NoError(t, err)
	assert.Equal(t, "pc_laptops", pclaptopsCat.Code())

}

func TestFactoryCanBuildConfigurableTest(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t, "products2.csv")

	product, err := rep.FindByMarketplaceCode(context.Background(), "CONF-1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "CONF-1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup Configurable")
	assert.Equal(t, "computers", product.BaseData().MainCategory.Code)

	assert.IsType(t, domain.ConfigurableProduct{}, product)

	configurable, _ := product.(domain.ConfigurableProduct)

	variant, err := configurable.Variant("1000000")
	assert.NoError(t, err, "Expected Variant with code 1000000 under configurable")
	assert.Equal(t, variant.MarketPlaceCode, "1000000", "wrong marketplacecode in variant")

	assert.Contains(t, configurable.VariantVariationAttributes, "clothingSize", "wrong VariantVariationAttributes")
}

func TestPageSize(t *testing.T) {
	pageSizeA := 3
	pageSizeB := 6

	rep := getRepositoryWithFixturesLoaded(t, "products2.csv")

	pageSizeFilterA := searchDomain.NewPaginationPageSizeFilter(pageSizeA)
	productHits, err := rep.Find(context.Background(), pageSizeFilterA)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))
	assert.Equal(t, pageSizeA, len(productHits.Hits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeA, len(productHits.Hits)))

	pageSizeFilterB := searchDomain.NewPaginationPageSizeFilter(pageSizeB)
	productHits, err = rep.Find(context.Background(), pageSizeFilterB)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	assert.Equal(t, pageSizeB, len(productHits.Hits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeB, len(productHits.Hits)))
}

func TestSortDirection(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t, "products2.csv")

	ascendingFilter := searchDomain.NewSortFilter("name", "A")
	productHits, err := rep.Find(context.Background(), ascendingFilter)
	assert.NotNil(t, productHits)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	var resultsAsc []string

	for _, hit := range productHits.Hits {
		if hit.BaseData().HasAttribute("name") {
			resultsAsc = append(resultsAsc, string(hit.BaseData().Attributes["name"].Value()))
		}
	}

	assert.True(t, sort.StringsAreSorted(resultsAsc), "Values are not sorted")

	descendingFilter := searchDomain.NewSortFilter("name", "D")
	productHits, err = rep.Find(context.Background(), descendingFilter)
	assert.NotNil(t, productHits)
	assert.True(t, len(productHits.Hits) > 0, "expected at least a hit")

	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	var resultsDesc []string

	for _, hit := range productHits.Hits {
		if hit.BaseData().HasAttribute("name") {
			resultsDesc = append(resultsDesc, string(hit.BaseData().Attributes["name"].Value()))
		}
	}

	assert.NotNil(t, productHits)
	assert.True(t, len(productHits.Hits) > 0, "expected at least a hit")

	assert.Equal(t, reverseStringSlice(resultsAsc), resultsDesc, "Value order was not reversed")
}

func TestFilterByAttribute(t *testing.T) {
	attributeName := "20000733_lactoseFreeClaim"
	attributeValue := "30002654_yes"

	rep := getRepositoryWithFixturesLoaded(t, "products2.csv")

	attributeFilter := searchDomain.NewKeyValueFilter(attributeName, []string{attributeValue})
	productHits, err := rep.Find(context.Background(), attributeFilter)
	assert.NoError(t, err)
	assert.NotNil(t, productHits)
	assert.True(t, len(productHits.Hits) > 0, "expected at least a hit")
	for _, hit := range productHits.Hits {
		assert.Equal(t, attributeValue, hit.BaseData().Attributes[attributeName].Value())
	}
}

func reverseStringSlice(stringSlice []string) []string {
	last := len(stringSlice) - 1
	for i := 0; i < len(stringSlice)/2; i++ {
		stringSlice[i], stringSlice[last-i] = stringSlice[last-i], stringSlice[i]
	}

	return stringSlice
}

func getAppDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	fmt.Printf("Filename : %q, Dir : %q\n", filename, path.Dir(filename))

	return path.Dir(filename)
}

func getRepositoryWithFixturesLoaded(t *testing.T, productCsv string) *productsearch.InMemoryProductRepository {
	rep := &productsearch.InMemoryProductRepository{}
	indexer := &domain2.Indexer{}
	indexer.Inject(
		flamingo.NullLogger{},
		rep,
		&struct {
			CategoryRepository domain2.CategoryRepository `inject:",optional"`
		}{
			CategoryRepository: rep,
		},
	)
	loader := csvcommerceLoader.IndexUpdater{}
	loader.Inject(flamingo.NullLogger{},
		&domain2.CategoryTreeBuilder{},
		&struct {
			ProductCsvFile           string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.path"`
			ProductCsvDelimiter      string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.delimiter"`
			ProductAttributesToSplit config.Slice `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.attributesToSplit"`
			CategoryCsvFile          string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.path,optional"`
			CategoryCsvDelimiter     string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.delimiter,optional"`
			Locale                   string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.locale"`
			Currency                 string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.currency"`
		}{
			Currency:             "GBP",
			Locale:               "en_GB",
			ProductCsvFile:       "fixture/" + productCsv,
			CategoryCsvFile:      "fixture/categories.csv",
			ProductCsvDelimiter:  ",",
			CategoryCsvDelimiter: ",",
		},
	)
	err := loader.Index(context.Background(), indexer)
	assert.NoError(t, err)
	return rep
}

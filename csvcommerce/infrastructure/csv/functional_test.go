package csv_test

import (
	csvcommerceLoader "flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/infrastructure/productSearch"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"fmt"
	"sort"
	"testing"

	"path"
	"runtime"

	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"github.com/stretchr/testify/assert"
)

func TestFactoryCanBuildSimpleTest(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t)


	product, err := rep.FindByMarketplaceCode("1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup")
}

func TestFactoryCanBuildConfigurableTest(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t)

	product, err := rep.FindByMarketplaceCode("CONF-1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "CONF-1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup Configurable")

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


	rep := getRepositoryWithFixturesLoaded(t)

	pageSizeFilterA := searchDomain.NewPaginationPageSizeFilter(pageSizeA)
	productHits, err := rep.Find(pageSizeFilterA)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))
	assert.Equal(t, pageSizeA, len(productHits.Hits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeA, len(productHits.Hits)))

	pageSizeFilterB := searchDomain.NewPaginationPageSizeFilter(pageSizeB)
	productHits, err = rep.Find(pageSizeFilterB)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	assert.Equal(t, pageSizeB, len(productHits.Hits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeB, len(productHits.Hits)))
}

func TestSortDirection(t *testing.T) {

	rep := getRepositoryWithFixturesLoaded(t)

	ascendingFilter := searchDomain.NewSortFilter("name", "A")
	productHits, err := rep.Find(ascendingFilter)
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
	productHits, err = rep.Find(descendingFilter)
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


	rep := getRepositoryWithFixturesLoaded(t)

	attributeFilter := searchDomain.NewKeyValueFilter(attributeName, []string{attributeValue})
	productHits, err := rep.Find(attributeFilter)
	assert.NoError(t,err)
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


func getRepositoryWithFixturesLoaded(t *testing.T) productSearch.ProductRepository{
	rep := &productSearch.InMemoryProductRepository{}
	loader := csvcommerceLoader.Loader{}
	loader.Inject(flamingo.NullLogger{},
		&struct {
			CsvFile  string `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
			Locale   string `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.locale"`
			Currency string `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.currency"`
		}{
			Currency: "GBP",
			Locale:   "en_GB",
			CsvFile:  "fixture/products2.csv",
		},
	)
	err := loader.Load(rep)
	assert.NoError(t, err)
	return rep
}
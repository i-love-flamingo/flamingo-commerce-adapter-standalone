package infrastructure_test

import (
	"testing"
	"fmt"
	"os"
	"sort"

	"github.com/stretchr/testify/assert"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	"flamingo.me/flamingo-commerce/product/domain"
	searchDomain "flamingo.me/flamingo-commerce/search/domain"
)

func TestFactoryCanBuildSimpleTest(t *testing.T) {
	factory := productRepository.InMemoryProductRepositoryFactory{}

	currentDir, _ := os.Getwd()
	rep, err := factory.BuildFromProductCSV(currentDir+"/../../csvCommerce/infrastructure/csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)

	product, err := rep.FindByMarketplaceCode("1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup")
}

func TestFactoryCanBuildConfigurableTest(t *testing.T) {
	factory := productRepository.InMemoryProductRepositoryFactory{}

	currentDir, _ := os.Getwd()
	rep, err := factory.BuildFromProductCSV(currentDir+"/../../csvCommerce/infrastructure/csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)

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

	factory := productRepository.InMemoryProductRepositoryFactory{}

	currentDir, _ := os.Getwd()
	rep, err := factory.BuildFromProductCSV(currentDir+"/../../csvCommerce/infrastructure/csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)

	pageSizeFilterA := searchDomain.NewPaginationPageSizeFilter(pageSizeA)
	productHits, err := rep.Find(pageSizeFilterA)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))
	assert.Equal(t, pageSizeA, len(productHits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeA, len(productHits)))

	pageSizeFilterB := searchDomain.NewPaginationPageSizeFilter(pageSizeB)
	productHits, err = rep.Find(pageSizeFilterB)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))
	assert.Equal(t, pageSizeB, len(productHits), fmt.Sprintf("Expected to get %d results but got %d", pageSizeB, len(productHits)))
}

func TestSortDirection(t *testing.T) {
	factory := productRepository.InMemoryProductRepositoryFactory{}

	currentDir, _ := os.Getwd()
	rep, err := factory.BuildFromProductCSV(currentDir+"/../../csvCommerce/infrastructure/csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)

	ascendingFilter := searchDomain.NewSortFilter("name", "A")
	productHits, err := rep.Find(ascendingFilter)
	assert.NotNil(t, productHits)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	var resultsAsc []string

	for _, hit := range productHits {
		if hit.BaseData().HasAttribute("name") {
			resultsAsc = append(resultsAsc, string(hit.BaseData().Attributes["name"].Value()))
		}
	}

	assert.True(t, sort.StringsAreSorted(resultsAsc), "Values are not sorted")

	descendingFilter := searchDomain.NewSortFilter("name", "D")
	productHits, err = rep.Find(descendingFilter)
	assert.NotNil(t, productHits)
	assert.NoError(t, err, fmt.Sprintf("Finding Products resulted in an error %s", err))

	var resultsDesc []string

	for _, hit := range productHits {
		if hit.BaseData().HasAttribute("name") {
			resultsDesc = append(resultsDesc, string(hit.BaseData().Attributes["name"].Value()))
		}
	}

	assert.NotNil(t, productHits)
	assert.Equal(t, reverseStringSlice(resultsAsc), resultsDesc, "Value order was not reversed")
}

func TestFilterByAttribute(t *testing.T) {
	attributeName := "20000733_lactoseFreeClaim"
	attributeValue := "30002654_yes"

	factory := productRepository.InMemoryProductRepositoryFactory{}

	currentDir, _ := os.Getwd()
	rep, err := factory.BuildFromProductCSV(currentDir+"/../../csvCommerce/infrastructure/csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)

	attributeFilter := searchDomain.NewKeyValueFilter(attributeName, []string{attributeValue})
	productHits, err := rep.Find(attributeFilter)
	assert.NotNil(t, productHits)

	for _, hit := range productHits {
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

package productRepository_test

import (
	"testing"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	"flamingo.me/flamingo-commerce/product/domain"
	"github.com/stretchr/testify/assert"
)

func TestFactoryCanBuildSimpleTest(t *testing.T) {
	factory := productRepository.InMemoryProductRepositoryFactory{}
	rep, err := factory.BuildFromProductCSV("../csv/fixture/products.csv", "en_GB", "GBP")
	assert.NoError(t, err)
	product, err := rep.FindByMarketplaceCode("1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup")
}

func TestFactoryCanBuildConfigurableTest(t *testing.T) {
	factory := productRepository.InMemoryProductRepositoryFactory{}
	rep, err := factory.BuildFromProductCSV("../csv/fixture/products.csv", "en_GB", "GBP")
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

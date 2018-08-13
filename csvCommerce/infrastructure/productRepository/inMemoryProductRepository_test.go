package productRepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	factory := InMemoryProductRepositoryFactory{}
	rep, err := factory.BuildFromProductCSV("../csv/fixture/products.csv", "en_GB")
	assert.NoError(t, err)
	product, err := rep.FindByMarketplaceCode("1000000")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, product.BaseData().MarketPlaceCode, "1000000")
	assert.Equal(t, product.BaseData().Title, "Hello Kitty Candy Cup")
}

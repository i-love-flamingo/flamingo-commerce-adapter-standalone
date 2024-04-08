//go:build integration
// +build integration

package moduletest_test

import (
	"net/http"
	"testing"

	"flamingo.me/dingo"

	"flamingo.me/flamingo/v3/framework/config"

	"flamingo.me/flamingo-commerce/v3/test/integrationtest"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing"
)

func testEmptySearchReturnsAllProducts(info integrationtest.BootupInfo) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
		expect := e.GET("/search").Expect()
		productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
		productsResult.Value("Hits").Array().Length().Equal(50)
		searchMeta := productsResult.Value("SearchMeta").Object()
		searchMeta.Value("NumResults").Number().Equal(67)
		searchMeta.Value("NumPages").Number().Equal(2)
		searchMeta.Value("Page").Number().Equal(1)
	}
}

func testSearchWithQueryReturnsAllMatchingProducts(info integrationtest.BootupInfo) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
		expect := e.GET("/search").WithQueryString("q=Refrigerator").Expect()
		productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
		productsResult.Value("Hits").Array().Length().Equal(1)
		product := productsResult.Value("Hits").Array().First().Object()
		product.Value("Title").String().Equal("Lucky Refrigerator Kidline TITLE")
		product.Value("ShortDescription").String().Equal("The Nellie swung to her anchor.")
		product.Value("Description").String().Equal("The Nellie, a cruising yawl, swung to her anchor without a flutter of the sails, and was at rest.The sun set; the dusk fell on the stream, and lights began to appear along the shore. The Chapman light–house, a three–legged thing erect on a mud–flat, shone strongly.The Nellie, a cruising yawl, swung to her anchor without a flutter of the sails, and was at rest.")
		product.Value("MarketPlaceCode").String().Equal("awesome-retailer_9468736")
		product.Value("RetailerCode").String().Equal("awesome-retailer")
		product.Value("StockLevel").String().Equal("in")
		product.Value("IsSaleable").Boolean().Equal(true)
		product.Value("Attributes").Object().Value("brandCode").Object().Value("Label").String().Equal("shouldermire")
		product.Value("Attributes").Object().Value("brandCode").Object().Value("Code").String().Equal("brandCode")
		product.Value("Attributes").Object().Value("gtin").Object().Value("Label").String().Equal("189905108817")
		product.Value("Attributes").Object().Value("gtin").Object().Value("Code").String().Equal("gtin")
		product.Value("Media").Array().Length().Equal(3)
		product.Value("Keywords").Array().ContainsOnly("keyword1", "keyword2")
		product.Value("ActivePrice").Object().Value("Default").Object().Value("Amount").String().Equal("132.31")
		product.Value("ActivePrice").Object().Value("Discounted").Object().Value("Amount").String().Equal("115.00")
		product.Value("ActivePrice").Object().Value("Default").Object().Value("Currency").String().Equal("GBP")
		product.Value("ActivePrice").Object().Value("Discounted").Object().Value("Currency").String().Equal("GBP")
		product.Value("ActivePrice").Object().Value("IsDiscounted").Boolean().Equal(true)
		searchMeta := productsResult.Value("SearchMeta").Object()
		searchMeta.Value("NumResults").Number().Equal(1)
		searchMeta.Value("NumPages").Number().Equal(1)
		searchMeta.Value("Page").Number().Equal(1)
	}
}

func testSearchWithPaginationWorks(info integrationtest.BootupInfo) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
		expect := e.GET("/search").WithQueryString("page=2").Expect()
		productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
		productsResult.Value("Hits").Array().Length().Equal(17)
		searchMeta := productsResult.Value("SearchMeta").Object()
		searchMeta.Value("NumResults").Number().Equal(67)
		searchMeta.Value("NumPages").Number().Equal(2)
		searchMeta.Value("Page").Number().Equal(2)
	}
}

func Test_BleveAdapter(t *testing.T) {
	bleveFlamingo := integrationtest.Bootup(
		[]dingo.Module{
			new(commercesearch.Module),
			new(commercesearch.CategoryModule),
			new(commercesearch.SearchModule),
			new(csvindexing.ProductModule),
		},
		"",
		config.Map{
			"flamingoCommerceAdapterStandalone.commercesearch.repositoryAdapter": "bleve",
		},
	)
	t.Cleanup(bleveFlamingo.ShutdownFunc)

	t.Run("bleve: Empty search returns all results", testEmptySearchReturnsAllProducts(bleveFlamingo))
	t.Run("bleve: Search with query works", testSearchWithQueryReturnsAllMatchingProducts(bleveFlamingo))
	t.Run("bleve: Search with pagination works", testSearchWithPaginationWorks(bleveFlamingo))
}

func Test_InMemoryAdapter(t *testing.T) {
	inMemoryFlamingo := integrationtest.Bootup(
		[]dingo.Module{
			new(commercesearch.Module),
			new(commercesearch.CategoryModule),
			new(commercesearch.SearchModule),
			new(csvindexing.ProductModule),
		},
		"",
		config.Map{
			"flamingoCommerceAdapterStandalone.commercesearch.repositoryAdapter": "inmemory",
		},
	)
	t.Cleanup(inMemoryFlamingo.ShutdownFunc)

	t.Run("inmemory: Empty search returns all results", testEmptySearchReturnsAllProducts(inMemoryFlamingo))
	t.Run("inmemory: Search with query works", testSearchWithQueryReturnsAllMatchingProducts(inMemoryFlamingo))
	t.Run("inmemory: Search with pagination works", testSearchWithPaginationWorks(inMemoryFlamingo))
}

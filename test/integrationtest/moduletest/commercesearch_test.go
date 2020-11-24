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

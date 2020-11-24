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

func Test_BleveAndInMemoryAdapter(t *testing.T) {
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
	defer bleveFlamingo.ShutdownFunc()

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
	defer inMemoryFlamingo.ShutdownFunc()

	adapters := map[string]integrationtest.BootupInfo{"inmemory": inMemoryFlamingo, "belve": bleveFlamingo}

	for adapter, info := range adapters {
		t.Run(adapter+":Empty search returns all results", func(t *testing.T) {
			t.Parallel()
			e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
			expect := e.GET("/search").Expect()
			productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
			productsResult.Value("Hits").Array().Length().Equal(50)
			searchMeta := productsResult.Value("SearchMeta").Object()
			searchMeta.Value("NumResults").Number().Equal(67)
			searchMeta.Value("NumPages").Number().Equal(2)
			searchMeta.Value("Page").Number().Equal(1)
		})

		t.Run(adapter+":Search with query returns all matching results", func(t *testing.T) {
			t.Parallel()
			e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
			expect := e.GET("/search").WithQueryString("q=Refrigerator").Expect()
			productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
			productsResult.Value("Hits").Array().Length().Equal(1)
			searchMeta := productsResult.Value("SearchMeta").Object()
			searchMeta.Value("NumResults").Number().Equal(1)
			searchMeta.Value("NumPages").Number().Equal(1)
			searchMeta.Value("Page").Number().Equal(1)
		})

		t.Run(adapter+":Search with pagination works", func(t *testing.T) {
			t.Parallel()
			e := integrationtest.NewHTTPExpect(t, "http://"+info.BaseURL)
			expect := e.GET("/search").WithQueryString("page=2").Expect()
			productsResult := expect.Status(http.StatusOK).JSON().Object().Value("SearchResult").Object().Value("products").Object()
			productsResult.Value("Hits").Array().Length().Equal(17)
			searchMeta := productsResult.Value("SearchMeta").Object()
			searchMeta.Value("NumResults").Number().Equal(67)
			searchMeta.Value("NumPages").Number().Equal(2)
			searchMeta.Value("Page").Number().Equal(2)
		})
	}
}

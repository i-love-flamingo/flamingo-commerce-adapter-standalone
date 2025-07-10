package main

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"flamingo.me/flamingo/v3/core/requestlogger"
	"flamingo.me/flamingo/v3/framework/prefixrouter"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing"
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder"
)

func main() {
	flamingo.App([]dingo.Module{
		new(requestlogger.Module),
		new(prefixrouter.Module),

		new(commercesearch.Module),
		new(commercesearch.CategoryModule),
		new(commercesearch.SearchModule),
		new(csvindexing.ProductModule),
		new(emailplaceorder.Module),
	})
}

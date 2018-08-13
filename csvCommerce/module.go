package csvCommerce

import (
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	categorydomain "flamingo.me/flamingo-commerce/category/domain"
	productdomain "flamingo.me/flamingo-commerce/product/domain"
	searchdomain "flamingo.me/flamingo-commerce/search/domain"
	"flamingo.me/flamingo/framework/dingo"
)

// ensure types for the Ports and Adapters
var (
	_ productdomain.ProductService   = nil
	_ productdomain.SearchService    = nil
	_ searchdomain.SearchService     = nil
	_ categorydomain.CategoryService = nil
)

type (

	// ProductClientModule for product client stuff
	ProductClientModule struct{}

	// SearchClientModule for searching
	SearchClientModule struct{}

	// CategoryClientModule for searching
	CategoryClientModule struct{}
)

// Configure DI
func (module *ProductClientModule) Configure(injector *dingo.Injector) {
	injector.Bind((*productdomain.ProductService)(nil)).To(infrastructure.ProductServiceAdapter{})

	injector.Bind((*productRepository.InMemoryProductRepository)(nil)).ToProvider(func(provider *productRepository.InMemoryProductRepositoryProvider) *productRepository.InMemoryProductRepository {
		rep, err := provider.GetForCurrentLocale()
		if err != nil {
			panic("cannot get InMemoryProductRepository")
		}
		return rep
	}).In(dingo.ChildSingleton)
}

// Configure DI
func (module *SearchClientModule) Configure(injector *dingo.Injector) {
	/*injector.Bind((*searchdomain.SearchService)(nil)).To(search.Service{})
	injector.Bind((*productdomain.SearchService)(nil)).To(search.ProductSearchServiceAdapter{})

	injector.BindMap(search.Decoder(nil), "product").ToInstance(search.ProductDecoder)

	// Bind specific Search Endpoints
	injector.BindMap(search.Endpoint(nil), "search").ToInstance(search.EndpointConfigForBaseSearch)

	// Bind specific Search Endpoints for product search service
	injector.BindMap(search.Endpoint(nil), "products_search").ToInstance(search.EndpointConfigForBaseProductSearch)
	injector.BindMap(search.Endpoint(nil), "products_by_category").ToInstance(search.EndpointConfigForCategory)
	injector.BindMap(search.Endpoint(nil), "products_by_retailer").ToInstance(search.EndpointConfigForProductsByRetailer)
	injector.BindMap(search.Endpoint(nil), "products_by_brand").ToInstance(search.EndpointConfigForProductsByBrand)

	//injector.BindMap(search.Decoder(nil), "brand").ToInstance(search.BrandDemoDecoder)
	*/
}

// Configure DI
func (module *CategoryClientModule) Configure(injector *dingo.Injector) {
	//injector.Bind((*categorydomain.CategoryService)(nil)).To(category.Service{})
}

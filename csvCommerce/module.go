package csvCommerce

import (
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/interfaces/controller"
	inMemoryProductSearchInfrastructure "flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	categorydomain "flamingo.me/flamingo-commerce/category/domain"
	productdomain "flamingo.me/flamingo-commerce/product/domain"
	searchdomain "flamingo.me/flamingo-commerce/search/domain"
	"flamingo.me/flamingo/framework/dingo"
	"flamingo.me/flamingo/framework/router"
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

	injector.Bind((*inMemoryProductSearchInfrastructure.InMemoryProductRepository)(nil)).ToProvider(
		func(provider *productRepository.InMemoryProductRepositoryProvider) *inMemoryProductSearchInfrastructure.InMemoryProductRepository {
			rep, err := provider.GetForCurrentLocale()
			if err != nil {
				panic("cannot get InMemoryProductRepository")
			}
			return rep
		}).In(dingo.ChildSingleton)

	router.Bind(injector, new(routes))
}

// Configure DI
func (module *SearchClientModule) Configure(injector *dingo.Injector) {
	injector.Bind((*productdomain.SearchService)(nil)).To(infrastructure.ProductSearchServiceAdapter{})
	/*
		injector.Bind((*searchdomain.SearchService)(nil)).To(search.ProductSearchServiceAdapter{})

		injector.BindMap(search.Decoder(nil), "product").ToInstance(search.ProductDecoder)

	*/
}

// Configure DI
func (module *CategoryClientModule) Configure(injector *dingo.Injector) {
	//injector.Bind((*categorydomain.CategoryService)(nil)).To(category.Service{})
}

type routes struct {
	controller *controller.ImageController
}

func (r *routes) Inject(controller *controller.ImageController) {
	r.controller = controller
}

func (r *routes) Routes(registry *router.Registry) {
	registry.HandleGet("csvcommerce.image.get", r.controller.Get)
	registry.Route("/image/:size/:filename", `csvcommerce.image.get(size,filename)`)
}

package csvcommerce

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/infrastructure"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/infrastructure/productrepository"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/interfaces/controller"
	inMemoryProductSearchInfrastructure "flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	categorydomain "flamingo.me/flamingo-commerce/v3/category/domain"
	productdomain "flamingo.me/flamingo-commerce/v3/product/domain"
	searchdomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/web"
)

// Ensure types for the Ports and Adapters
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
		func(provider *productrepository.InMemoryProductRepositoryBuilder) *inMemoryProductSearchInfrastructure.InMemoryProductRepository {
			rep, err := provider.GetForCurrentLocale()
			if err != nil {
				panic("cannot get InMemoryProductRepository")
			}
			return rep
		}).In(dingo.ChildSingleton)

	web.BindRoutes(injector, new(routes))
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

func (r *routes) Routes(registry *web.RouterRegistry) {
	registry.HandleGet("csvcommerce.image.get", r.controller.Get)
	registry.Route("/image/:size/:filename", `csvcommerce.image.get(size,filename)`)
}

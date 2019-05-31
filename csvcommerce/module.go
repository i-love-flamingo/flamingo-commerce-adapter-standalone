package csvcommerce

import (
	"flamingo.me/dingo"
	productSearch2 "flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/infrastructure/productSearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/interfaces/controller"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"

	productSearchModule "flamingo.me/flamingo-commerce-adapter-standalone/productSearch"
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
	// ProductModule for product stuff
	ProductModule struct{}

)

// Configure DI
func (module *ProductModule) Configure(injector *dingo.Injector) {
	//Register Loader for productSearch
	injector.Bind((*productSearch.Loader)(nil)).To(productSearch2.Loader{}).In(dingo.ChildSingleton)

	web.BindRoutes(injector, new(routes))
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


// Depends on other modules
func (m *ProductModule) Depends() []dingo.Module {
	return []dingo.Module{
		new(productSearchModule.Module),
	}
}
package csvindexing

import (
	"flamingo.me/dingo"
	commercesearchModule "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch"
	commercesearchDomain "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/infrastructure/commercesearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/interfaces/controller"
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
func (m *ProductModule) Configure(injector *dingo.Injector) {
	//Register IndexUpdater for productSearch
	injector.Bind((*commercesearchDomain.IndexUpdater)(nil)).To(commercesearch.IndexUpdater{})

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
		new(commercesearchModule.Module),
	}
}

// CueConfig defines the cart module configuration
func (m *ProductModule) CueConfig() string {
	return `
flamingoCommerceAdapterStandalone: {
	csvindexing: {
		productCsvPath: string | *"ressources/products/products.csv"
		categoryCsvPath: string | *""
		locale: string | *"en"
		currency: string | *"€"
		allowedImageResizeParamaters: string | *"200x,300x,400x,x200,x300"
	}
}`
}

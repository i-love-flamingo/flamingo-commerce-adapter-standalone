package productSearch

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/category"
	"flamingo.me/flamingo-commerce/v3/category/domain"

	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/product"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"
	productdomain "flamingo.me/flamingo-commerce/v3/product/domain"

)


type (
	// Module for product client stuff
	Module struct{}

	// CategoryModule registers the Category Adapter that uses the productRepositry
	CategoryModule struct{}


)

// Configure DI
func (module *Module) Configure(injector *dingo.Injector) {
	injector.Bind((*productdomain.ProductService)(nil)).To(product.ServiceAdapter{})
	injector.Bind((*productdomain.SearchService)(nil)).To(product.SearchServiceAdapter{})

	injector.Bind((*productSearch.ProductRepository)(nil)).ToProvider(
		func(builder *productSearch.InMemoryProductRepositoryBuilder) *productSearch.InMemoryProductRepository {
			rep, err := builder.Get()
			if err != nil {
				panic("cannot get InMemoryProductRepository:"+err.Error())
			}
			return rep
		}).In(dingo.ChildSingleton)

}


// Configure DI
func (module *CategoryModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(domain.CategoryService)).To(category.Adapter{})
}


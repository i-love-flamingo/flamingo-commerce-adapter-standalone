package commercesearch

import (
	"context"

	"flamingo.me/dingo"
	commerceCategoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	commerceProduct "flamingo.me/flamingo-commerce/v3/product"
	commerceProductDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	commerceSearchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"flamingo.me/flamingo/v3/framework/web"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/category"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/product"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/productsearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/search"
)

type (
	// Module for product client stuff
	Module struct {
		repositoryAdapter string
	}

	// EventSubscriber for starting the index processes
	EventSubscriber struct {
		logger       flamingo.Logger
		indexProcess *domain.IndexProcess
	}

	// CategoryModule registers the Category Adapter that uses the productRepositry
	CategoryModule struct{}

	// SearchModule registers the Category Adapter that uses the productRepositry
	SearchModule struct{}
)

// Inject for subscriber
func (s *EventSubscriber) Inject(logger flamingo.Logger, indexProcess *domain.IndexProcess) {
	s.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone.commercesearch").WithField(flamingo.LogKeyCategory, "eventsubscriber")
}

// Notify should get called by flamingo event logic
func (s *EventSubscriber) Notify(ctx context.Context, event flamingo.Event) {
	// we want to start an Indexing Process for every routed AreaRoutedEvent
	if e, ok := event.(*web.AreaRoutedEvent); ok {
		s.logger.WithContext(ctx).Info("AreaRoutedEvent for Area:" + e.ConfigArea.Name)
		injector, err := e.ConfigArea.GetInitializedInjector()
		if err != nil {
			s.logger.Error(err)
			return
		}
		i, err := injector.GetInstance(domain.IndexProcess{})
		if err != nil {
			panic(err)
		}
		indexProcess := i.(*domain.IndexProcess)
		err = indexProcess.Run(ctx)
		if err != nil {
			s.logger.Error(err)
		}
	}
}

// Inject  module
func (m *Module) Inject(config *struct {
	RepositoryAdapter string `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.repositoryAdapter,optional"`
}) {
	if config != nil {
		m.repositoryAdapter = config.RepositoryAdapter
	}
}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	injector.Bind((*commerceProductDomain.ProductService)(nil)).To(product.ServiceAdapter{})
	injector.Bind((*commerceProductDomain.SearchService)(nil)).To(product.SearchServiceAdapter{})
	flamingo.BindEventSubscriber(injector).To(new(EventSubscriber))

	switch m.repositoryAdapter {
	case "bleve":
		injector.Bind((*domain.ProductRepository)(nil)).To(productsearch.BleveRepository{}).In(dingo.ChildSingleton)
		injector.Bind((*domain.CategoryRepository)(nil)).To(productsearch.BleveRepository{}).In(dingo.ChildSingleton)
	default:
		injector.Bind((*domain.ProductRepository)(nil)).To(productsearch.InMemoryProductRepository{}).In(dingo.ChildSingleton)
		injector.Bind((*domain.CategoryRepository)(nil)).To(productsearch.InMemoryProductRepository{}).In(dingo.ChildSingleton)
	}
}

// Depends on other modules
func (m *Module) Depends() []dingo.Module {
	return []dingo.Module{
		new(commerceProduct.Module),
	}
}

// Configure DI
func (module *CategoryModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(commerceCategoryDomain.CategoryService)).To(category.Adapter{})
}

// Configure DI
func (module *SearchModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(commerceSearchDomain.SearchService)).To(search.ServiceAdapter{})
}

// CueConfig defines the cart module configuration
func (*Module) CueConfig() string {
	return `
flamingoCommerceAdapterStandalone: {
	commercesearch: {
		enableIndexing: bool | *false
		repositoryAdapter: "bleve" | *"inmemory"
		bleveAdapter: {
			productsToParentCategories: bool | *true
			enableCategoryFacet: bool | *false
			facetConfig: [...{attributeCode: string, amount: number}]
			sortConfig:[...{attributeCode: string, attributeType: "numeric"|"bool"|*"text", asc: bool, desc: bool}]
		}
	}
}`
}

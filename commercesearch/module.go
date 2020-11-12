package commercesearch

import (
	"context"

	"flamingo.me/dingo"
	"flamingo.me/flamingo-commerce/v3/category/domain"
	domain2 "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"flamingo.me/flamingo/v3/framework/web"

	productSearchDomain "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/category"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/search"

	productdomain "flamingo.me/flamingo-commerce/v3/product/domain"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/product"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/productsearch"
)

type (
	// Module for product client stuff
	Module struct {
		repositoryAdapter string
	}

	//EventSubscriber for starting the index processes
	EventSubscriber struct {
		logger       flamingo.Logger
		indexProcess *productSearchDomain.IndexProcess
	}

	// CategoryModule registers the Category Adapter that uses the productRepositry
	CategoryModule struct{}

	// SearchModule registers the Category Adapter that uses the productRepositry
	SearchModule struct{}
)

// Inject for subscriber
func (s *EventSubscriber) Inject(logger flamingo.Logger, indexProcess *productSearchDomain.IndexProcess) {
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
		i, err := injector.GetInstance(productSearchDomain.IndexProcess{})
		if err != nil {
			panic(err)
		}
		indexProcess := i.(*productSearchDomain.IndexProcess)
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
	injector.Bind((*productdomain.ProductService)(nil)).To(product.ServiceAdapter{})
	injector.Bind((*productdomain.SearchService)(nil)).To(product.SearchServiceAdapter{})
	flamingo.BindEventSubscriber(injector).To(new(EventSubscriber))

	switch m.repositoryAdapter {
	case "bleve":
		injector.Bind((*productSearchDomain.ProductRepository)(nil)).To(productsearch.BleveRepository{}).In(dingo.ChildSingleton)
		injector.Bind((*productSearchDomain.CategoryRepository)(nil)).To(productsearch.BleveRepository{}).In(dingo.ChildSingleton)

		/*
			injector.Bind((*productSearchDomain.ProductRepository)(nil)).ToProvider(
				func(logger flamingo.Logger, repo *productSearch.BleveRepository, indexProcess *productSearchDomain.IndexProcess, config *struct {
					EnableIndexing bool `inject:"config:flamingoCommerceAdapterStandalone.productSearch.enableIndexing,optional"`
				}) productSearchDomain.ProductRepository {
					enableIndexing := false
					if config != nil {
						enableIndexing = config.EnableIndexing
					}
					return indexIntoRepo(logger, repo, indexProcess, enableIndexing)
				}).In(dingo.ChildSingleton)
		*/

	default:
		injector.Bind((*productSearchDomain.ProductRepository)(nil)).To(productsearch.InMemoryProductRepository{}).In(dingo.ChildSingleton)
		injector.Bind((*productSearchDomain.CategoryRepository)(nil)).To(productsearch.InMemoryProductRepository{}).In(dingo.ChildSingleton)

		/*
			injector.Bind((*productSearchDomain.ProductRepository)(nil)).ToProvider(
				func(logger flamingo.Logger, repo *productSearch.InMemoryProductRepository, indexProcess *productSearchDomain.IndexProcess, config *struct {
					EnableIndexing bool `inject:"config:flamingoCommerceAdapterStandalone.productSearch.enableIndexing,optional"`
				}) productSearchDomain.ProductRepository {
					enableIndexing := false
					if config != nil {
						enableIndexing = config.EnableIndexing
					}
					return indexIntoRepo(logger, repo, indexProcess, enableIndexing)
				}).In(dingo.ChildSingleton)
		*/

	}
}

// Configure DI
func (module *CategoryModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(domain.CategoryService)).To(category.Adapter{})
}

// Configure DI
func (module *SearchModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(domain2.SearchService)).To(search.ServiceAdapter{})
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

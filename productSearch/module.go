package productSearch

import (
	"flamingo.me/dingo"
	productSearchDomain "flamingo.me/flamingo-commerce-adapter-standalone/productSearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/category"
	"flamingo.me/flamingo-commerce/v3/category/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/product"
	"flamingo.me/flamingo-commerce-adapter-standalone/productSearch/infrastructure/productSearch"
	productdomain "flamingo.me/flamingo-commerce/v3/product/domain"

)


type (
	// Module for product client stuff
	Module struct{
		repositoryAdapter string
	}

	// CategoryModule registers the Category Adapter that uses the productRepositry
	CategoryModule struct{}
)

// Inject  module
func (m *Module) Inject(config *struct {
	RepositoryAdapter string `inject:"config:flamingo-commerce-adapter-standalone.repositoryAdapter,optional"`
}) {
	if config != nil {
		m.repositoryAdapter = config.RepositoryAdapter
	}
}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	injector.Bind((*productdomain.ProductService)(nil)).To(product.ServiceAdapter{})
	injector.Bind((*productdomain.SearchService)(nil)).To(product.SearchServiceAdapter{})

	switch m.repositoryAdapter {
	case "bleve":
		injector.Bind((*productSearchDomain.ProductRepository)(nil)).ToProvider(
			func(logger flamingo.Logger,repo *productSearch.BleveRepository, indexer *productSearchDomain.Indexer, config *struct {EnableIndexing bool `inject:"config:flamingo-commerce-adapter-standalone.enableIndexing,optional"`}) productSearchDomain.ProductRepository {
				enableIndexing := false
				if config != nil {
					enableIndexing = config.EnableIndexing
				}
				return indexIntoRepo(logger,repo,indexer,enableIndexing)
			}).In(dingo.ChildSingleton)
	default:
		injector.Bind((*productSearchDomain.ProductRepository)(nil)).ToProvider(
			func(logger flamingo.Logger,repo *productSearch.InMemoryProductRepository, indexer *productSearchDomain.Indexer, config *struct {EnableIndexing bool `inject:"config:flamingo-commerce-adapter-standalone.enableIndexing,optional"`}) productSearchDomain.ProductRepository {
				enableIndexing := false
				if config != nil {
					enableIndexing = config.EnableIndexing
				}
				return indexIntoRepo(logger,repo,indexer,enableIndexing)
			}).In(dingo.ChildSingleton)
	}


}

func indexIntoRepo(logger flamingo.Logger,repo productSearchDomain.ProductRepository, indexer *productSearchDomain.Indexer, enableIndexing bool) productSearchDomain.ProductRepository{
	if !enableIndexing {
		logger.WithField(flamingo.LogKeyModule,"flamingo-commerce-adapter-standalone").Info("Indexing disabled")
		return repo
	}
	logger.WithField(flamingo.LogKeyModule,"flamingo-commerce-adapter-standalone").Info("Indexing enabled in area")
	err := indexer.Fill(repo)
	if err != nil {
		logger.WithField(flamingo.LogKeyModule,"flamingo-commerce-adapter-standalone").Fatal("Cannot index",err)
	}
	return repo
}


// Configure DI
func (module *CategoryModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(domain.CategoryService)).To(category.Adapter{})
}


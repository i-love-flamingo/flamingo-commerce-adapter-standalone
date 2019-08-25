# product search Module

Provides a standalone search for products and Adapters for flamingo_commerce productService and productSearchService.

The basic idea is:
* index products only once (per configuration area) in a standalone "productrepository" - this instance is then used by the adapters
* the details of loading of products for indexing is NOT done by this module but should be done by other modules. 
For example there is a Loader that reads products from CSV in the module "csvCommerce"


## Loading of products

Loading of products is done by implementing and registering the Loader interface:
```go

	//Loader - interface to Load products in a Index
    Loader interface {
        Load(ctx context.Context,indexer *Indexer)
    }
```

and register it in:

```
injector.Bind((*productSearchDomain.Loader)(nil)).To(YourLoaderImplementation)
```

### Configuration

```

flamingo-commerce-adapter-standalone:
  enableIndexing: false
  repositoryAdapter: inmemory

```

Set the repositoryAdapter to "bleve" for an experimental repository implementation using https://github.com/blevesearch/bleve/ search
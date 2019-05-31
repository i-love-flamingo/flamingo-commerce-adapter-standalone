# product search Module

Provides a standalone search for products and Adapters for flamingo_commerce productService and productSearchService.

The basic idea is:
* index products only once (per configuration area) in a standalone "productrepository" - this instance is then used by the adapters
* the loading of products for indexing is NOT done by this module but should be done by other modules. For example there is a Loader that reads products from CSV.


## Loading of products

Loading of products is done by implementing and registering the Loader interface:
```go

	//Indexer - interface to index products
	Indexer interface {
		Add(ctx context.Context,product domain.BasicProduct) error
	}

	//Loader - interface to Load products in a Index
    Loader interface {
        Load(ctx context.Context,indexer *Indexer)
    }
```


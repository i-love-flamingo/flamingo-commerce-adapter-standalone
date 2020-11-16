# Commerce (Product) Search Module

Provides Adapters for Flamingo Commerce to persist and retrieve products.

The provided adapters are:

* productService (to retrieve single products)
* productSearchService (to search for products - e.g. used on category listing page)
* categoryService (for Flamingo Commerce "category" module - to receive categorys and category trees). To use the
  Adapter you need to add the main `CategoryModule` to your bootstrap.
* searchService (for Flamingo Commerce "search" module allowing searching for products documents). To use the Adapters
  you need to add the main `SearchModule` to your bootstrap.

The available products need to be indexed first and will be stored in a `ProductRepository`.

The module also provides an adapter to receive categories and Category Trees and provides an Adpater for the Flamingo
Commerce CategoryService (flamingo.me/flamingo-commerce/v3/category/domain)

## Indexing

The indexing itself is not part of that module - because the indexing source might be something project specific.
However - the module "csvindexing" implements an indexer that can be used to read products from a CSV file.

The indexing (loading) of products is done by implementing and registering the `IndexUpdater` interface:

```go

// IndexUpdater interface to Load products in a Index - secondary port
IndexUpdater interface {
// Indexer method that is called with an initialized Indexer. The passed Indexer provides helpers to update the Repository
Index(ctx context.Context, rep *Indexer) error
}
```

So you can implement that interface (port) by an own implementation (adapter) and then register your implementation:

```go
injector.Bind((*productSearchDomain.IndexUpdater)(nil)).To(YourLoaderImplementation)
```

## Configuration

With the setting
`flamingoCommerceAdapterStandalone.commercesearch.repositoryAdapter` you can switch the repository implementation.

The default is a simple in-memory product index, that works for single instances.

### Bleve Repository Adapter

You can also use the bleve based repository - bleve (http://blevesearch.com/) is a full text search and index for go.
Supports sorting, facets and more.

```yaml
flamingoCommerceAdapterStandalone:
  commercesearch:
    enableIndexing: true
    repositoryAdapter: "bleve"
    bleveAdapter:
      productsToParentCategories: true
      enableCategoryFacet: true
      facetConfig:
        # Add facet for the color attribute
        - attributeCode: "color"
          amount: 20
      sortConfig:
        # Add sorting for color attribute
        - attributeCode: "color"
          attributeType: "text" # field content: text, numeric, bool
          asc: true # Allow asc sorting
          desc: true # Allow desc sorting
```
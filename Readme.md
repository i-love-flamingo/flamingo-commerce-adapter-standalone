[![Tests](https://github.com/i-love-flamingo/flamingo-commerce-adapter-standalone/workflows/Tests/badge.svg?branch=master)](https://github.com/i-love-flamingo/flamingo-commerce-adapter-standalone/actions?query=workflow%3ATests+branch%3Amaster)

# Flamingo Commerce Adapters Standalone

This repository contains modules that allow to run Flamingo Commerce in a standalone mode.
(Not connected to any third party headless ecommerce API).

According to the Flamingo Commerce concept of "ports and adapters" the modules provide implementations for Flamingo Commerce Ports.

The following flamingo modules are part of that:

* **commercesearch**
  * Provide Product; ProductSearch; Category and Search. That means you can have working product views, category views and search features.
  * The module therefore internally uses Repository where products are stored and received from. You can choose between
    * A simple "InMemory" version
    * An implementation based on bleve - a go based indexed search implementation ( https://github.com/blevesearch/bleve)
  * The module requires someone who takes care of indexing new products and categories. Therefore, it expects an implementation of an "IndexUpdater"
  * See [Module commercesearch Readme](commercesearch/Readme.md)
    
* **csvindexing**
  * useful addon module for the *commercesearch* module
  * Provide a "IndexUpdater" that loads products and categories from a CSV file
  * See [Module csvcommerce Readme](csvindexing/Readme.md)
    
* **emailplaceorder**
  * Like the name says a module that just sends mails (to customer and store owner) after placing an order
  * See [Module emailplaceorder Readme](emailplaceorder/Readme.md)

    
## Usage

Just add the modules to your Flamingo bootstrap like this:

```go
		new(commercesearch.Module),
		new(commercesearch.CategoryModule),
		new(commercesearch.SearchModule),
		new(csvindexing.ProductModule),
		new(emailplaceorder.Module),
        
```

There are a couple of configuration options. See the Flamingo `config` command, and the module readme for details.

[![Build Status](https://travis-ci.org/i-love-flamingo/flamingo-commerce-adapter-standalone.svg?branch=master)](https://travis-ci.org/i-love-flamingo/flamingo-commerce-adapter-standalone?branch=master)


# Flamingo Commerce Adapters Standalone 

This repository contains modules that allow to run Flamingo Commerce n a standalone mode.
(Not connected to any thirds party headless ecommerc API).

According to the Flamingo Commerce concept of "ports and adapters" the modules provide mplementations for Flamingo Commerce Ports.

The following flamingo modules are part of that:

* commercesearch
    * Provide Product; ProductSearch; Category and Search. That means you can have working product views, category vews and search features.
    * The module therefore interally uses Repository where products are stored and received from. You can choose between
        * A simple "InMemory" version
        * A implementation based on bleve - a go based indexed search implementation ( https://github.com/blevesearch/bleve)
    * The module requires someone who takes care if indexing new products and categories. Therefore it expects an implemantation of an "IndexUpdater"
    * See [Module commercesearch Readme](commercesearch/Readme.md)
    
* csvindexing
    * usful addon module for the *commercesearch* module
    * Provide a "IndexUpdater" that loads products and categories from a CSV file
    * See [Module csvcommerce Readme](csvindexing/Readme.md)
    
* emailplaceorder
    * Like the name says a module that just sends mails (to customer and storeowner) after placing an order
    

    
## Usage

Just add the modules to your Flamingo bootstrap like this:

```
        //flamingo-commerce-adpater-standalone modules:
		new(commercesearch.Module),
		new(commercesearch.CategoryModule),
		new(commercesearch.SearchModule),
		new(csvindexing.ProductModule),
		new(emailplaceorder.Module),
        ..
```

There are a couple of configuration options. See the Flamingo `config` command and the module readme for details.
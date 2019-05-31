package productSearch

import (
	"flamingo.me/flamingo-commerce/v3/product/domain"
)

type (

	//Index - interface to index products
	Index interface {
		Add(product domain.BasicProduct) error
		FindByMarketplaceCode(vcode string) (domain.BasicProduct, error)
	}

	//Loader - interface to Load products in a Index
	Loader interface {
		Load(indexer Index) error
	}
)

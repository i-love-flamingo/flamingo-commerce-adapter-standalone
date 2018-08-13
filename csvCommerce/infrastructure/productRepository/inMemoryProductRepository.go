package productRepository

import (
	"flamingo.me/flamingo-commerce/product/domain"
	"flamingo.me/flamingo/framework/flamingo"
)

type (
	InMemoryProductRepository struct {
		products map[string]domain.BasicProduct
	}

	InMemoryProductRepositoryProvider struct {
		InMemoryProductRepositoryFactory *InMemoryProductRepositoryFactory `inject:""`
		Logger                           flamingo.Logger                   `inject:""`
		Locale                           string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.locale"`
		ProductCSVPath                   string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
	}
)

var buildedRepositoryByLocale = make(map[string]*InMemoryProductRepository)

// Get returns a product struct
func (r *InMemoryProductRepository) add(product domain.BasicProduct) error {
	if r.products == nil {
		r.products = make(map[string]domain.BasicProduct)
	}
	r.products[product.BaseData().MarketPlaceCode] = product
	return nil
}

// Get returns a product struct
func (r *InMemoryProductRepository) FindByMarketplaceCode(marketplaceCode string) (domain.BasicProduct, error) {
	if v, ok := r.products[marketplaceCode]; ok {
		return v, nil
	}
	return nil, domain.ProductNotFound{
		MarketplaceCode: marketplaceCode,
	}
}

func (p *InMemoryProductRepositoryProvider) GetForCurrentLocale() (*InMemoryProductRepository, error) {
	locale := p.Locale
	if v, ok := buildedRepositoryByLocale[locale]; ok {
		return v, nil
	}
	p.Logger.Info("Build InMemoryProductRepository for locale " + locale + " .....")
	rep, err := p.InMemoryProductRepositoryFactory.BuildFromProductCSV(p.ProductCSVPath, locale)
	if err != nil {
		return nil, err
	}
	buildedRepositoryByLocale[locale] = rep
	return buildedRepositoryByLocale[locale], nil
}

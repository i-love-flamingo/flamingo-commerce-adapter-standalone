package productrepository

import (
	"sync"

	inMemoryProductSearchInfrastructure "flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	"flamingo.me/flamingo/v3/framework/flamingo"
)

type (
	// InMemoryProductRepositoryBuilder uses the factory to build a new InMemoryProductRepository - or uses the cached version
	InMemoryProductRepositoryBuilder struct {
		InMemoryProductRepositoryFactory *InMemoryProductRepositoryFactory `inject:""`
		Logger                           flamingo.Logger                   `inject:""`
		Locale                           string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.locale"`
		Currency                         string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.currency"`
		ProductCSVPath                   string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
		rwMutex                          sync.RWMutex
	}
)

var buildedRepositoryByLocale = make(map[string]*inMemoryProductSearchInfrastructure.InMemoryProductRepository)

// GetForCurrentLocale returns a Product Repository of the In Memory Type prefiltered for a given locale
func (p *InMemoryProductRepositoryBuilder) GetForCurrentLocale() (*inMemoryProductSearchInfrastructure.InMemoryProductRepository, error) {
	locale := p.Locale

	p.rwMutex.RLock()
	if v, ok := buildedRepositoryByLocale[locale]; ok {
		defer p.rwMutex.RUnlock()
		return v, nil
	}
	p.rwMutex.RUnlock()

	p.Logger.Info("Build InMemoryProductRepository for locale " + locale + " .....")

	rep, err := p.InMemoryProductRepositoryFactory.BuildFromProductCSV(p.ProductCSVPath, locale, p.Currency)
	if err != nil {
		return nil, err
	}
	p.rwMutex.Lock()
	buildedRepositoryByLocale[locale] = rep
	p.rwMutex.Unlock()

	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	return buildedRepositoryByLocale[locale], nil
}

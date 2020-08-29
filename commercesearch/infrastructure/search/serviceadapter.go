package search

import (
	"context"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
)

type (
	//ServiceAdapter for search service
	ServiceAdapter struct {
		productRepository domain.ProductRepository
	}
)

var _ searchDomain.SearchService = &ServiceAdapter{}

//Inject - dingo injector
func (p *ServiceAdapter) Inject(productRepository domain.ProductRepository) {
	p.productRepository = productRepository
}

//Search implementation
func (p *ServiceAdapter) Search(ctx context.Context, filter ...searchDomain.Filter) (map[string]searchDomain.Result, error) {
	res, err := p.productRepository.Find(ctx, filter...)
	if err != nil {
		return nil, err
	}
	result := searchDomain.Result{
		SearchMeta: res.Result.SearchMeta,
		Suggestion: res.Suggestion,
		Facets:     res.Facets,
	}
	var hits []searchDomain.Document
	for _, h := range res.Hits {
		hits = append(hits, searchDomain.Document(h))
	}
	result.Hits = hits
	mapRes := make(map[string]searchDomain.Result)
	mapRes["products"] = result
	return mapRes, nil
}

//SearchFor implementation
func (p *ServiceAdapter) SearchFor(ctx context.Context, typ string, filter ...searchDomain.Filter) (*searchDomain.Result, error) {
	res, err := p.Search(ctx, filter...)
	if err != nil {
		return nil, err
	}
	psr := res["products"]
	return &psr, nil
}

package productSearch

import (

	"flamingo.me/flamingo/v3/framework/flamingo"
	"sync"
)

type (
	// InMemoryProductRepositoryBuilder uses the factory to build a new InMemoryProductRepository - or uses the cached version
	InMemoryProductRepositoryBuilder struct {
		Loader         Loader         `inject:""`
		Logger         flamingo.Logger `inject:""`

	}
)

var mutex       sync.Mutex

func (p *InMemoryProductRepositoryBuilder) Get() (*InMemoryProductRepository, error) {

	mutex.Lock()
	defer mutex.Unlock()

	newRepo := InMemoryProductRepository{}


	err := p.Loader.Load(&newRepo)
	if err != nil {
		return nil, err
	}

	return &newRepo, nil
}

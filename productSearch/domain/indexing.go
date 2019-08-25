package domain

import (
	"flamingo.me/flamingo/v3/framework/flamingo"
	"sync"
)

type (
	// Indexer - responsible to call the injected loader to index products into the passed repository
	Indexer struct {
		loader Loader
		logger flamingo.Logger
	}

	//Loader - interface to Load products in a Index - secondary port
	Loader interface {
		Load(rep ProductRepository) error
	}
)

var mutex sync.Mutex

func (p *Indexer) Inject(loader Loader, logger flamingo.Logger) {
	p.loader = loader
	p.logger = logger.WithField(flamingo.LogKeyModule,"flamingo-commerce-adapter-standalone").WithField(flamingo.LogKeyCategory,"indexer")
}

func (p *Indexer) Fill(rep ProductRepository) (error) {

	mutex.Lock()
	defer mutex.Unlock()

	p.logger.Info("Prepareing Index..")
	err := rep.PrepareIndex()
	if err != nil {
		return err
	}

	p.logger.Info("Start registered Loader..")
	err = p.loader.Load(rep)
	if err != nil {
		return err
	}

	return nil
}

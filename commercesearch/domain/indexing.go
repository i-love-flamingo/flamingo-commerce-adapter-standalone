package domain

import (
	"context"
	"fmt"
	"sync"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	product "flamingo.me/flamingo-commerce/v3/product/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
)

type (
	// IndexProcess responsible to call the injected loader to index products into the passed repository
	IndexProcess struct {
		indexUpdater   IndexUpdater
		indexer        *Indexer
		logger         flamingo.Logger
		enableIndexing bool
	}

	// Indexer provides useful features to work with the Repositories for indexing purposes
	Indexer struct {
		productRepository  ProductRepository
		categoryRepository CategoryRepository
		logger             flamingo.Logger
		batchProductQueue  []product.BasicProduct
		batchCatQueue      []product.CategoryTeaser
	}

	// CategoryTreeBuilder helper to build category tree
	CategoryTreeBuilder struct {
		// rootCategory - this is the link into the tree that is going to be build
		rootCategory *categoryDomain.TreeData

		// categoryTreeIndex - the link into the treenode - is build
		categoryTreeIndex map[string]*categoryDomain.TreeData

		// child -> parent
		nodeLinkRawData map[string]string
	}

	categoryRawNode struct {
		code   string
		name   string
		parent string
	}

	// IndexUpdater - interface to update the index with the help of the Indexer
	IndexUpdater interface {
		Index(ctx context.Context, rep *Indexer) error
	}
)

var mutex sync.Mutex

// Inject for Indexer
func (i *Indexer) Inject(logger flamingo.Logger, productRepository ProductRepository,
	config *struct {
		CategoryRepository CategoryRepository `inject:",optional"`
	}) *Indexer {
	i.logger = logger
	i.productRepository = productRepository
	if config != nil {
		i.categoryRepository = config.CategoryRepository
	}

	return i
}

// PrepareIndex of the avaiable repository implementations
func (i *Indexer) PrepareIndex(ctx context.Context) error {
	err := i.productRepository.PrepareIndex(ctx)
	if err != nil {
		return err
	}
	if i.categoryRepository != nil {
		return i.categoryRepository.PrepareIndex(ctx)
	}
	return nil
}

// ProductRepository to get
func (i *Indexer) ProductRepository() ProductRepository {
	return i.productRepository
}

func (i *Indexer) commit(ctx context.Context) error {
	err := i.productRepository.UpdateProducts(ctx, i.batchProductQueue)
	if err != nil {
		return err
	}
	i.batchProductQueue = nil
	if i.categoryRepository != nil {
		err = i.categoryRepository.UpdateByCategoryTeasers(ctx, i.batchCatQueue)
		if err != nil {
			return err
		}
		i.batchCatQueue = nil
	}
	return nil
}

// UpdateProductAndCategory helper to update product and the assigned categoryteasers
func (i *Indexer) UpdateProductAndCategory(ctx context.Context, product product.BasicProduct) error {
	i.batchProductQueue = append(i.batchProductQueue, product)

	if product.BaseData().Categories != nil {
		i.batchCatQueue = append(i.batchCatQueue, product.BaseData().Categories...)
	}

	if product.BaseData().MainCategory.Code != "" {
		i.batchCatQueue = append(i.batchCatQueue, product.BaseData().MainCategory)
	}
	return i.commit(ctx)
}

// Inject dependencies
func (p *IndexProcess) Inject(indexUpdater IndexUpdater, logger flamingo.Logger, indexer *Indexer, config *struct {
	EnableIndexing bool `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.enableIndexing,optional"`
}) {
	p.indexUpdater = indexUpdater
	p.indexer = indexer
	p.enableIndexing = config.EnableIndexing
	p.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone").WithField(flamingo.LogKeyCategory, "indexer")
}

// Run the index process with registered loader (using indexer as helper for the repository access)
func (p *IndexProcess) Run(ctx context.Context) error {
	if !p.enableIndexing {
		p.logger.Info("Skipping Indexing..")
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()

	p.logger.Info("Prepareing Indexes..")
	err := p.indexer.PrepareIndex(ctx)
	if err != nil {
		return err
	}

	p.logger.Info("Start registered Indexer..")
	err = p.indexUpdater.Index(ctx, p.indexer)
	if err != nil {
		return err
	}

	p.logger.Info("Indexing finished..")

	return nil
}

// AddCategoryData to the builder.. Call this as often as you want to add before calling BuildTree
func (h *CategoryTreeBuilder) AddCategoryData(code string, name string, parentCode string) {
	if h.categoryTreeIndex == nil {
		h.categoryTreeIndex = make(map[string]*categoryDomain.TreeData)
	}
	if h.rootCategory == nil {
		h.rootCategory = &categoryDomain.TreeData{}
	}
	// root category is either a default empty node or detected by code == parentCode
	if code == parentCode {
		h.rootCategory = &categoryDomain.TreeData{
			CategoryCode: code,
			CategoryName: name,
		}
		h.categoryTreeIndex[code] = h.rootCategory
		return
	}

	if h.nodeLinkRawData == nil {
		h.nodeLinkRawData = make(map[string]string)
	}
	builtBasicNode := categoryDomain.TreeData{
		CategoryCode: code,
		CategoryName: name,
	}
	h.categoryTreeIndex[code] = &builtBasicNode
	h.nodeLinkRawData[code] = parentCode
}

// BuildTree build Tree based on added categoriedata
func (h *CategoryTreeBuilder) BuildTree() (*categoryDomain.TreeData, error) {
	// Build the tree links
	for childCode, parentCode := range h.nodeLinkRawData {
		childNode, ok := h.categoryTreeIndex[childCode]
		if !ok {
			return nil, fmt.Errorf("ChildNode %v not found", childNode)
		}
		var parentNode *categoryDomain.TreeData
		if parentCode == "" {
			parentNode = h.rootCategory
		} else {
			parentNode, ok = h.categoryTreeIndex[parentCode]
			if !ok {
				return nil, fmt.Errorf("ParentCode %v not found", parentCode)
			}
		}
		parentNode.SubTreesData = append(parentNode.SubTreesData, childNode)
	}
	buildPathString(h.rootCategory)
	return h.rootCategory, nil
}

func buildPathString(parent *categoryDomain.TreeData) {
	// Build the Path
	for _, subNode := range parent.SubTreesData {
		subNode.CategoryPath = parent.CategoryPath + "/" + subNode.CategoryCode
		buildPathString(subNode)
	}
}

// CategoryTreeToCategoryTeaser conversion
func CategoryTreeToCategoryTeaser(searchedCategoryCode string, tree categoryDomain.Tree) *product.CategoryTeaser {
	return categoryTreeToCategoryTeaser(searchedCategoryCode, tree, nil)
}

func categoryTreeToCategoryTeaser(searchedCategoryCode string, searchPosition categoryDomain.Tree, parentCategory *product.CategoryTeaser) *product.CategoryTeaser {
	teaserForCurrentNode := &product.CategoryTeaser{
		Code:   searchPosition.Code(),
		Path:   searchPosition.Path(),
		Name:   searchPosition.Name(),
		Parent: parentCategory,
	}
	// recursion stops of category found
	if searchPosition.Code() == searchedCategoryCode {
		return teaserForCurrentNode
	}
	for _, subNode := range searchPosition.SubTrees() {
		found := categoryTreeToCategoryTeaser(searchedCategoryCode, subNode, teaserForCurrentNode)
		if found != nil {
			return found
		}
	}
	return nil
}

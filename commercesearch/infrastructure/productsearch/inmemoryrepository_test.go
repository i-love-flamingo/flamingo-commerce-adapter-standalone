package productsearch

import (
	"context"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/require"
	"testing"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryProductRepository_AddProduct(t *testing.T) {
	s := &InMemoryProductRepository{
		logger: flamingo.NullLogger{},
	}

	product := domain.SimpleProduct{
		Identifier: "id",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id",
			Title:           "a title",
			MainCategory: domain.CategoryTeaser{
				Code: "Sub",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
	}
	product2 := domain.SimpleProduct{
		Identifier: "id2",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id2",
			Title:           "b title",
			MainCategory: domain.CategoryTeaser{
				Code: "Sub",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
	}
	err := s.UpdateProducts(context.Background(), []domain.BasicProduct{product, product2})
	assert.NoError(t, err)

	found, _ := s.FindByMarketplaceCode(context.Background(), "id")
	assert.Equal(t, "a title", found.BaseData().Title)

	result, _ := s.Find(context.Background(), categoryDomain.CategoryFacet{
		CategoryCode: "Sub",
	}, searchDomain.NewSortFilter("title", "A"))
	require.Len(t, result.Hits, 2, "expect 2 results for category search")
	assert.Equal(t, "a title", result.Hits[0].BaseData().Title)
	assert.Equal(t, "b title", result.Hits[1].BaseData().Title)

	//test pagination
	t.Run("Test pagination", func(t *testing.T) {
		result, _ = s.Find(context.Background(), categoryDomain.CategoryFacet{
			CategoryCode: "Sub",
		}, searchDomain.NewPaginationPageFilter(1), searchDomain.NewSortFilter("title", "A"))

		require.Len(t, result.Hits, 2, "expect 2 results for category search on page 1")
		assert.Equal(t, "a title", result.Hits[0].BaseData().Title)

		result, _ = s.Find(context.Background(),
			categoryDomain.CategoryFacet{
				CategoryCode: "Sub",
			},
			searchDomain.NewPaginationPageFilter(2),
			searchDomain.NewPaginationPageSizeFilter(10),
			searchDomain.NewSortFilter("title", "A"),
		)
		assert.Len(t, result.Hits, 0, "expected no results on page 2")

		result, _ = s.Find(context.Background(),
			categoryDomain.CategoryFacet{
				CategoryCode: "Sub",
			},
			searchDomain.NewPaginationPageFilter(2),
			searchDomain.NewPaginationPageSizeFilter(1),
			searchDomain.NewSortFilter("title", "A"),
		)
		require.Len(t, result.Hits, 1)
		assert.Equal(t, "b title", result.Hits[0].BaseData().Title)
	})

}

func TestInMemoryProductRepository_CategorySearch(t *testing.T) {
	s := &InMemoryProductRepository{}
	s.UpdateByCategoryTeasers(context.Background(), []domain.CategoryTeaser{
		domain.CategoryTeaser{
			Code: "Sub",
			Parent: &domain.CategoryTeaser{
				Code: "Root",
			},
		},
		domain.CategoryTeaser{
			Code: "pc_laptops",
			Parent: &domain.CategoryTeaser{
				Code: "computers",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
	})

	cat, _ := s.Category(context.Background(), "")
	if assert.NotNil(t, cat) {
		assert.Equal(t, "Root", cat.Code())
	}

	cat, _ = s.Category(context.Background(), "Sub")
	if assert.NotNil(t, cat) {
		assert.Equal(t, "Sub", cat.Code())
	}
	cat, _ = s.Category(context.Background(), "pc_laptops")
	if assert.NotNil(t, cat) {
		assert.Equal(t, "pc_laptops", cat.Code())
	}

	tree, _ := s.CategoryTree(context.Background(), "")
	if assert.NotNil(t, tree) {
		assert.Equal(t, "Root", tree.Code())
		assert.Equal(t, "Sub", tree.SubTrees()[0].Code())
	}
}

func TestInMemoryProductRepository_categoryTeaserToCategoryTree(t *testing.T) {
	s := &InMemoryProductRepository{}
	teaser := domain.CategoryTeaser{
		Code: "subsub",
		Parent: &domain.CategoryTeaser{
			Code: "sub",
			Parent: &domain.CategoryTeaser{
				Code: "root",
			},
		},
	}
	result := s.categoryTeaserToCategoryTree(teaser, nil)
	assert.Equal(t, "root", result.CategoryCode)
	assert.Equal(t, "sub", result.SubTreesData[0].CategoryCode)
}

func TestInMemoryProductRepository_addCategoryPath(t *testing.T) {

	existingTree := &categoryDomain.TreeData{
		CategoryCode: "root",
		SubTreesData: []*categoryDomain.TreeData{
			&categoryDomain.TreeData{
				CategoryCode: "sub1",
			},
			&categoryDomain.TreeData{
				CategoryCode: "sub2",
			},
		},
	}

	mergeInTree := &categoryDomain.TreeData{
		CategoryCode: "root",
		SubTreesData: []*categoryDomain.TreeData{
			&categoryDomain.TreeData{
				CategoryCode: "sub3",
				SubTreesData: []*categoryDomain.TreeData{
					&categoryDomain.TreeData{
						CategoryCode: "sub3-sub",
					},
				},
			},
		},
	}
	mergeInTree2 := &categoryDomain.TreeData{
		CategoryCode: "root",
		SubTreesData: []*categoryDomain.TreeData{
			&categoryDomain.TreeData{
				CategoryCode: "sub2",
				SubTreesData: []*categoryDomain.TreeData{
					&categoryDomain.TreeData{
						CategoryCode: "sub2-sub",
					},
				},
			},
		},
	}
	s := &InMemoryProductRepository{}
	s.addCategoryPath(existingTree, mergeInTree)
	s.addCategoryPath(existingTree, mergeInTree2)

	assert.Equal(t, "root", existingTree.CategoryCode)

	assert.Equal(t, "sub1", existingTree.SubTreesData[0].CategoryCode)
	assert.Equal(t, "sub2", existingTree.SubTreesData[1].CategoryCode)
	assert.Equal(t, "sub2-sub", existingTree.SubTreesData[1].SubTreesData[0].CategoryCode)

	assert.Equal(t, "sub3", existingTree.SubTreesData[2].CategoryCode)
	assert.Equal(t, "sub3-sub", existingTree.SubTreesData[2].SubTreesData[0].CategoryCode)

}

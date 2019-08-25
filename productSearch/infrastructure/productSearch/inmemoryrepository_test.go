package productSearch

import (
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
			Title: "title",
			MainCategory:domain.CategoryTeaser{
				Code: "Sub",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
	}
	err := s.Add(product)
	assert.NoError(t,err)

	err = s.Add(product)

	found, _ := s.FindByMarketplaceCode("id")
	assert.Equal(t, "title", found.BaseData().Title)

	cat, _ := s.Category("")
	if assert.NotNil(t,cat) {
		assert.Equal(t, "Root", cat.Code())
	}

	cat, _ = s.Category("Sub")
	if assert.NotNil(t,cat) {
		assert.Equal(t, "Sub", cat.Code())
	}

	result, _ := s.Find(categoryDomain.CategoryFacet{
		CategoryCode:"Sub",
	})
	require.Len(t,result.Hits,1)
	assert.Equal(t,"title",result.Hits[0].BaseData().Title)

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
	s.addCategoryPath(existingTree,mergeInTree)
	s.addCategoryPath(existingTree,mergeInTree2)

	assert.Equal(t, "root",existingTree.CategoryCode)

	assert.Equal(t, "sub1",existingTree.SubTreesData[0].CategoryCode)
	assert.Equal(t, "sub2",existingTree.SubTreesData[1].CategoryCode)
	assert.Equal(t, "sub2-sub",existingTree.SubTreesData[1].SubTreesData[0].CategoryCode)

	assert.Equal(t, "sub3",existingTree.SubTreesData[2].CategoryCode)
	assert.Equal(t, "sub3-sub",existingTree.SubTreesData[2].SubTreesData[0].CategoryCode)

}

package commercesearch

import (
	"context"
	"math/big"
	"testing"
	"time"

	commercePriceDomain "flamingo.me/flamingo-commerce/v3/price/domain"
	searchDomain "flamingo.me/flamingo-commerce/v3/search/domain"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/require"

	categoryDomain "flamingo.me/flamingo-commerce/v3/category/domain"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	"github.com/stretchr/testify/assert"
)

func TestBleveProductRepository_AddProduct(t *testing.T) {
	s := &BleveRepository{}
	s.Inject(flamingo.NullLogger{}, nil)
	err := s.PrepareIndex(context.Background())
	assert.NoError(t, err)

	product := domain.SimpleProduct{
		Identifier: "id",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id",
			Title:           "atitle",
			MainCategory: domain.CategoryTeaser{
				Code: "Sub1",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
			Categories: []domain.CategoryTeaser{
				domain.CategoryTeaser{
					Code: "Sub3",
					Parent: &domain.CategoryTeaser{
						Code: "Root",
					},
				},
			},
		},
		Saleable: domain.Saleable{},
		Teaser: domain.TeaserData{
			ShortTitle:       "atitle",
			ShortDescription: "",
			URLSlug:          "",
			TeaserPrice: domain.PriceInfo{
				Default:      commercePriceDomain.NewFromInt(999, 100, "€"),
				Discounted:   commercePriceDomain.NewFromInt(899, 100, "€"),
				IsDiscounted: true,
			},
			TeaserPriceIsFromPrice:   false,
			PreSelectedVariantSku:    "",
			Media:                    nil,
			MarketPlaceCode:          "",
			TeaserAvailablePrices:    nil,
			TeaserLoyaltyPriceInfo:   nil,
			TeaserLoyaltyEarningInfo: nil,
		},
	}
	product2 := domain.SimpleProduct{
		Identifier: "id2",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id2",
			Title:           "btitle",
			MainCategory: domain.CategoryTeaser{
				Code: "Sub2",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
		Saleable: domain.Saleable{},
		Teaser: domain.TeaserData{
			ShortTitle:       "btitle",
			ShortDescription: "",
			URLSlug:          "",
			TeaserPrice: domain.PriceInfo{
				Default:      commercePriceDomain.NewFromInt(799, 100, "€"),
				IsDiscounted: false,
			},
			TeaserPriceIsFromPrice:   false,
			PreSelectedVariantSku:    "",
			Media:                    nil,
			MarketPlaceCode:          "",
			TeaserAvailablePrices:    nil,
			TeaserLoyaltyPriceInfo:   nil,
			TeaserLoyaltyEarningInfo: nil,
		},
	}

	product3 := domain.SimpleProduct{
		Identifier: "id3",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id3",
			Title:           "green bag of something",
			MainCategory: domain.CategoryTeaser{
				Code: "Sub2",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
		Saleable: domain.Saleable{},
		Teaser: domain.TeaserData{
			ShortTitle:       "green bag of something",
			ShortDescription: "Hello kitty!",
			URLSlug:          "",
			TeaserPrice: domain.PriceInfo{
				Default:      commercePriceDomain.NewFromInt(799, 100, "€"),
				IsDiscounted: false,
			},
			TeaserPriceIsFromPrice:   false,
			PreSelectedVariantSku:    "",
			Media:                    nil,
			MarketPlaceCode:          "",
			TeaserAvailablePrices:    nil,
			TeaserLoyaltyPriceInfo:   nil,
			TeaserLoyaltyEarningInfo: nil,
		},
	}
	err = s.UpdateProducts(context.Background(), []domain.BasicProduct{product, product2, product3})
	assert.NoError(t, err)

	// test pagination
	t.Run("Find product by Id", func(t *testing.T) {
		found, err := s.FindByMarketplaceCode(context.Background(), "id")
		require.NoError(t, err)
		assert.Equal(t, "atitle", found.BaseData().Title)

	})

	// test if category was indexed as well
	t.Run("Find by different Queries", func(t *testing.T) {
		result, _ := s.Find(context.Background(), searchDomain.NewQueryFilter("atitle"), searchDomain.NewSortFilter("Title", "A"))
		require.Len(t, result.Hits, 1, "expect 1 results for 'atitle' search")

		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("id2"), searchDomain.NewSortFilter("Title", "A"))
		require.Len(t, result.Hits, 1, "expect 1 results for 'id2' search")

		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("*"), searchDomain.NewSortFilter("Title", "A"))
		require.Equal(t, 3, result.SearchMeta.NumResults, "expect 3 results for * search")
		assert.Equal(t, "atitle", result.Hits[0].BaseData().Title, "ascending should have a first")

		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("kitty"), searchDomain.NewSortFilter("Title", "A"))
		require.Equal(t, 1, result.SearchMeta.NumResults, "expect 1 results for 'kitty' search")
		assert.Equal(t, "green bag of something", result.Hits[0].BaseData().Title, "ascending should have a first")

	})

	t.Run("Find with Sorting", func(t *testing.T) {
		result, _ := s.Find(context.Background(), searchDomain.NewQueryFilter("*"), searchDomain.NewSortFilter("price", "A"))
		require.Equal(t, 3, result.SearchMeta.NumResults, "expect 2 results for * search")
		assert.Equal(t, 7.99, result.Hits[0].TeaserData().TeaserPrice.GetFinalPrice().FloatAmount(), "ascending price should have cheapest price first")
		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("*"), searchDomain.NewSortFilter("price", "D"))
		assert.Equal(t, 8.99, result.Hits[0].TeaserData().TeaserPrice.GetFinalPrice().FloatAmount(), "descending price should have expensivst first")
	})

	// test if category was indexed as well
	t.Run("Filter by category", func(t *testing.T) {
		result, _ := s.Find(context.Background(), searchDomain.NewQueryFilter("*"), categoryDomain.NewCategoryFacet("Sub1"))
		require.Equal(t, 1, result.SearchMeta.NumResults, "expect 1 results for 'Sub1' category search")
		assert.Equal(t, "atitle", result.Hits[0].BaseData().Title)

		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("*"), categoryDomain.NewCategoryFacet("Sub2"))
		require.Equal(t, 2, result.SearchMeta.NumResults, "expect 1 results for 'Sub2' category search")

		result, _ = s.Find(context.Background(), searchDomain.NewQueryFilter("*"), categoryDomain.NewCategoryFacet("Sub3"))
		require.Equal(t, 1, result.SearchMeta.NumResults, "expect 1 results for 'Sub2' category search")

	})

	// test pagination
	t.Run("Test pagination", func(t *testing.T) {

		result, _ := s.Find(context.Background(), searchDomain.NewQueryFilter("*"), searchDomain.NewPaginationPageFilter(1), searchDomain.NewPaginationPageSizeFilter(2), searchDomain.NewSortFilter("Title", "A"))

		require.Len(t, result.Hits, 2, "expect 2 results for category search on page 1")
		require.Equal(t, 2, result.SearchMeta.NumPages, "expect 2 pages")
		require.Equal(t, 3, result.SearchMeta.NumResults, "expect 2 rsulsts in total")
		require.Equal(t, 1, result.SearchMeta.Page, "expect to be on 1 page")
		assert.Equal(t, "atitle", result.Hits[0].BaseData().Title)

	})

}

func TestBleveRepository_FacetsSearch(t *testing.T) {
	s := &BleveRepository{}

	configuration := config.Slice{}
	configuration = append(configuration, config.Map{
		"attributeCode": "brand",
		"amount":        10,
	})
	s.Inject(flamingo.NullLogger{}, &struct {
		AssignProductsToParentCategories bool         `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.productsToParentCategories,optional"`
		EnableCategoryFacet              bool         `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.enableCategoryFacet,optional"`
		FacetConfig                      config.Slice `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.facetConfig"`
		SortConfig                       config.Slice `inject:"config:flamingoCommerceAdapterStandalone.commercesearch.bleveAdapter.sortConfig"`
	}{
		AssignProductsToParentCategories: true,
		EnableCategoryFacet:              true,
		FacetConfig:                      configuration,
	})
	err := s.PrepareIndex(context.Background())
	assert.NoError(t, err)

	product := domain.SimpleProduct{
		Identifier: "id",

		BasicProductData: domain.BasicProductData{
			Title:           "atitle",
			MarketPlaceCode: "id",
			Attributes:      make(domain.Attributes),
			Categories: []domain.CategoryTeaser{
				domain.CategoryTeaser{
					Code: "Sub3",
					Parent: &domain.CategoryTeaser{
						Code: "Root",
					},
				},
			},
			MainCategory: domain.CategoryTeaser{
				Code: "Sub1",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
			CategoryToCodeMapping: nil,
			StockLevel:            "",
			Keywords:              nil,
			IsNew:                 false,
		},
		Saleable: domain.Saleable{},
	}
	product.BasicProductData.Attributes["brand"] = domain.Attribute{
		Code:      "brand",
		CodeLabel: "Brand",
		Label:     "apple",
		RawValue:  nil,
		UnitCode:  "",
	}
	product2 := domain.SimpleProduct{
		Identifier: "id2",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id2",
			Title:           "btitle",
			Attributes:      make(domain.Attributes),
			MainCategory: domain.CategoryTeaser{
				Code: "Sub2",
				Path: "",
				Name: "",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
		},
		Saleable: domain.Saleable{},
	}
	product2.BasicProductData.Attributes["brand"] = domain.Attribute{
		Code:      "brand",
		CodeLabel: "Brand",
		Label:     "apple",
		RawValue:  nil,
		UnitCode:  "",
	}

	product3 := domain.SimpleProduct{
		Identifier: "id3",
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode: "id3",
			Title:           "green bag of something",
			Attributes:      make(domain.Attributes),
			MainCategory: domain.CategoryTeaser{
				Code: "Sub2",
				Parent: &domain.CategoryTeaser{
					Code: "Root",
				},
			},
			Categories: []domain.CategoryTeaser{
				domain.CategoryTeaser{
					Code: "Sub1_Sub2",
					Parent: &domain.CategoryTeaser{
						Code: "Sub2",
						Parent: &domain.CategoryTeaser{
							Code: "Root",
						},
					},
				},
			},
		},
		Saleable: domain.Saleable{},
	}
	product3.BasicProductData.Attributes["brand"] = domain.Attribute{
		Code:      "brand",
		CodeLabel: "Brand",
		Label:     "flamingo",
		RawValue:  nil,
		UnitCode:  "",
	}
	err = s.UpdateProducts(context.Background(), []domain.BasicProduct{product, product2, product3})
	assert.NoError(t, err)

	// test pagination
	t.Run("Test facets", func(t *testing.T) {

		result, _ := s.Find(context.Background(),
			searchDomain.NewQueryFilter("*"))

		require.Equal(t, 3, len(result.Hits))
		require.Greater(t, len(result.Facets), 1)

		hasCatFacet := false
		hasCatFacetSub2 := false
		hasCatFacetSub1Sub2 := false
		hasBrandFacet := false
		hasBrandFacetApple := false
		for facetKey, facet := range result.Facets {
			if facetKey == "category" {
				hasCatFacet = true
				for _, facetItem := range facet.Items {
					if facetItem.Value == "Sub2" {
						hasCatFacetSub2 = true
						for _, facetItemL2 := range facetItem.Items {
							if facetItemL2.Value == "Sub1_Sub2" {
								hasCatFacetSub1Sub2 = true
							}
						}
					}
				}
			}
			if facetKey == "brand" {
				hasBrandFacet = true
				for _, facetItem := range facet.Items {
					if facetItem.Label == "apple" {
						hasBrandFacetApple = true
					}
				}
			}
		}
		assert.True(t, hasCatFacet)
		assert.True(t, hasCatFacetSub2)
		assert.True(t, hasCatFacetSub1Sub2)
		assert.True(t, hasBrandFacet)
		assert.True(t, hasBrandFacetApple)
	})
}

func TestBleveRepository_CategorySearch(t *testing.T) {

	s := &BleveRepository{}
	s.Inject(flamingo.NullLogger{}, nil)
	err := s.PrepareIndex(context.Background())
	assert.NoError(t, err)

	s.UpdateByCategoryTeasers(context.Background(), []domain.CategoryTeaser{
		domain.CategoryTeaser{
			Code: "Sub",
			Path: "",
			Name: "Subname",
			Parent: &domain.CategoryTeaser{
				Code:   "Root",
				Path:   "",
				Name:   "",
				Parent: nil,
			},
		},
		domain.CategoryTeaser{
			Code: "SubSub2",
			Path: "",
			Name: "",
			Parent: &domain.CategoryTeaser{
				Code: "Sub2",
				Path: "",
				Name: "",
				Parent: &domain.CategoryTeaser{
					Code:   "Root",
					Path:   "",
					Name:   "",
					Parent: nil,
				},
			},
		},
	})
	// test if category was indexed as well
	t.Run("Find category code", func(t *testing.T) {
		cat, _ := s.Category(context.Background(), "Sub")
		if assert.NotNil(t, cat) {
			assert.Equal(t, "Sub", cat.Code())
			assert.Equal(t, "Subname", cat.Name())
		}

		cat, _ = s.Category(context.Background(), "Sub2")
		if assert.NotNil(t, cat) {
			assert.Equal(t, "Sub2", cat.Code())
		}

		cat, _ = s.Category(context.Background(), "")
		if assert.NotNil(t, cat) {
			assert.Equal(t, "Root", cat.Code())
		}

	})

	// test if category was indexed as well
	t.Run("Get category tree", func(t *testing.T) {
		cat, _ := s.CategoryTree(context.Background(), "")
		if assert.NotNil(t, cat) {
			assert.Equal(t, "Root", cat.Code())
			assert.True(t, cat.HasChilds())
			hasSub := false
			hasSub2 := false
			for _, sub := range cat.SubTrees() {
				if sub.Code() == "Sub" {
					hasSub = true
				}
				if sub.Code() == "Sub2" {
					hasSub2 = true
				}
			}
			assert.True(t, hasSub, "hasSub should be true")
			assert.True(t, hasSub2, "hasSub2 should be true")
		}
	})

}

func TestBleveRepository_ProductDecodeEncode(t *testing.T) {

	r := &BleveRepository{}

	// test pagination
	t.Run("Test simpl", func(t *testing.T) {
		simpleProduct := domain.SimpleProduct{
			Identifier: "testid",
			BasicProductData: domain.BasicProductData{
				Title:      "Title",
				Categories: nil,
				MainCategory: domain.CategoryTeaser{
					Code:   "testcategory",
					Path:   "",
					Name:   "",
					Parent: nil,
				},
				CategoryToCodeMapping: nil,
				StockLevel:            "",
				Keywords:              nil,
				IsNew:                 false,
			},
			Saleable: domain.Saleable{
				IsSaleable:   false,
				SaleableFrom: time.Time{},
				SaleableTo:   time.Time{},
				ActivePrice: domain.PriceInfo{
					Default:           commercePriceDomain.NewFromInt(1, 100, "€"),
					Discounted:        commercePriceDomain.NewFromInt(1, 100, "€"),
					DiscountText:      "",
					ActiveBase:        big.Float{},
					ActiveBaseAmount:  big.Float{},
					ActiveBaseUnit:    "",
					IsDiscounted:      false,
					CampaignRules:     nil,
					DenyMoreDiscounts: false,
					Context:           domain.PriceContext{},
					TaxClass:          "",
				},
				LoyaltyEarnings: nil,
			},
			Teaser: domain.TeaserData{},
		}
		bytes, err := r.encodeProduct(simpleProduct)
		assert.NoError(t, err)
		productGot, err := r.decodeProduct(bytes)
		assert.NoError(t, err)
		assert.Equal(t, simpleProduct, productGot)
	})

	// test pagination
	t.Run("Test configurable", func(t *testing.T) {
		product := domain.ConfigurableProduct{
			Identifier: "testid",
			BasicProductData: domain.BasicProductData{
				Title:      "Title",
				Categories: nil,
				MainCategory: domain.CategoryTeaser{
					Code:   "testcategory",
					Path:   "",
					Name:   "",
					Parent: nil,
				},
				CategoryToCodeMapping: nil,
				StockLevel:            "",
				Keywords:              nil,
				IsNew:                 false,
			},
			Teaser: domain.TeaserData{},
		}
		bytes, err := r.encodeProduct(product)
		assert.NoError(t, err)
		productGot, err := r.decodeProduct(bytes)
		assert.NoError(t, err)
		assert.Equal(t, product, productGot)
	})

}

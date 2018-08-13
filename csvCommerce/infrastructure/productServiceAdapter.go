package infrastructure

import (
	"context"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/productRepository"
	"flamingo.me/flamingo-commerce/product/domain"
)

type (
	// ProductService interface
	ProductServiceAdapter struct {
		InMemoryProductRepositoryProvider *productRepository.InMemoryProductRepositoryProvider `inject:""`
	}
)

var (
	brands = []string{
		"Apple",
		"Bose",
		"Dior",
		"Hugo Boss",
	}
)

// Get returns a product struct
func (ps *ProductServiceAdapter) Get(ctx context.Context, marketplaceCode string) (domain.BasicProduct, error) {
	rep, err := ps.InMemoryProductRepositoryProvider.GetForCurrentLocale()
	if err != nil {
		return nil, err
	}
	return rep.FindByMarketplaceCode(marketplaceCode)
}

//defer ctx.Profile("service", "get product "+foreignId)()
/*

	if marketplaceCode == "fake_configurable" {
		product := ps.getFakeConfigurable(marketplaceCode)
		product.RetailerCode = "om3CommonTestretailer"
		product.Title = "TypeConfigurable product"

		product.VariantVariationAttributes = []string{"color", "size"}

		// prepare departure / arrival attributes
		// add departure / arrival to attributes
		attributeValueList := make([]interface{}, 2)
		attributeValueList[0] = "departure"
		attributeValueList[1] = "arrival"

		variants := []struct {
			marketplacecode string
			title           string
			attributes      domain.Attributes
		}{
			{"shirt-white-s", "Shirt White S", domain.Attributes{
				"size":             domain.Attribute{RawValue: "S"},
				"color":            domain.Attribute{RawValue: "white"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
			{"shirt-red-s", "Shirt Red S", domain.Attributes{
				"size":             domain.Attribute{RawValue: "S"},
				"color":            domain.Attribute{RawValue: "red"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
			{"shirt-white-m", "Shirt White M", domain.Attributes{
				"size":             domain.Attribute{RawValue: "M"},
				"color":            domain.Attribute{RawValue: "white"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
			{"shirt-black-m", "Shirt Black M", domain.Attributes{
				"size":             domain.Attribute{RawValue: "M"},
				"color":            domain.Attribute{RawValue: "black"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
			{"shirt-black-l", "Shirt Black L", domain.Attributes{
				"size":             domain.Attribute{RawValue: "L"},
				"color":            domain.Attribute{RawValue: "black"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
			{"shirt-red-l", "Shirt Red L", domain.Attributes{
				"size":             domain.Attribute{RawValue: "L"},
				"color":            domain.Attribute{RawValue: "red"},
				"collectionOption": domain.Attribute{RawValue: attributeValueList},
			},
			},
		}

		for _, variant := range variants {
			simpleVariant := ps.fakeVariant(variant.marketplacecode)
			simpleVariant.Title = variant.title
			simpleVariant.Attributes = variant.attributes

			product.Variants = append(product.Variants, simpleVariant)
		}

		return product, nil
	}
	if marketplaceCode == "fake_simple" {
		return ps.FakeSimple(marketplaceCode, false, false), nil
	}
	return nil, errors.New("Code " + marketplaceCode + " Not implemented in FAKE: Only code 'fake_configurable' or 'fake_simple' should be used")
}

func (ps *ProductServiceAdapter) FakeSimple(marketplaceCode string, isNew bool, isExclusive bool) domain.SimpleProduct {
	product := domain.SimpleProduct{}
	product.Title = "TypeSimple product"
	ps.addBasicData(&product.BasicProductData)

	product.Saleable = domain.Saleable{
		IsSaleable:   true,
		SaleableTo:   time.Now().Add(time.Hour * time.Duration(1)),
		SaleableFrom: time.Now().Add(time.Hour * time.Duration(-1)),
	}

	product.ActivePrice = ps.getPrice(20.99+float64(rand.Intn(10)), 10.49+float64(rand.Intn(10)))
	product.MarketPlaceCode = marketplaceCode

	product.Teaser = domain.TeaserData{
		ShortDescription: product.ShortDescription,
		ShortTitle:       product.Title,
		Media:            product.Media,
		MarketPlaceCode:  product.MarketPlaceCode,
	}

	if isNew {
		product.BasicProductData.IsNew = true
	}

	if isExclusive {
		product.Attributes["exclusiveProduct"] = domain.Attribute{
			RawValue: "30002654_yes",
			Code:     "exclusiveProduct",
		}
	}
	product.Attributes["size"] = domain.Attribute{
		RawValue: "XS",
		Code:     "size",
	}
	return product
}

func (ps *ProductServiceAdapter) getFakeConfigurable(marketplaceCode string) domain.ConfigurableProduct {
	product := domain.ConfigurableProduct{}
	product.Title = "TypeSimple product"
	ps.addBasicData(&product.BasicProductData)
	product.MarketPlaceCode = marketplaceCode

	return product
}

func (ps *ProductServiceAdapter) fakeVariant(marketplaceCode string) domain.Variant {
	var simpleVariant domain.Variant
	simpleVariant.Attributes = domain.Attributes(make(map[string]domain.Attribute))

	ps.addBasicData(&simpleVariant.BasicProductData)

	simpleVariant.ActivePrice = ps.getPrice(30.99+float64(rand.Intn(10)), 20.49+float64(rand.Intn(10)))
	simpleVariant.MarketPlaceCode = marketplaceCode
	simpleVariant.IsSaleable = true

	return simpleVariant
}

func (ps *ProductServiceAdapter) addBasicData(product *domain.BasicProductData) {
	product.ShortDescription = "Short Description"
	product.Description = "Description"
	product.Media = append(product.Media, domain.Media{Type: "image-external", Reference: "http://dummyimage.com/1024x768/000/fff", Usage: "detail"})
	product.Media = append(product.Media, domain.Media{Type: "image-external", Reference: "http://dummyimage.com/1024x768/000/fff", Usage: "detail"})
	product.Media = append(product.Media, domain.Media{Type: "image-external", Reference: "http://dummyimage.com/1024x768/000/fff", Usage: "detail"})
	product.Media = append(product.Media, domain.Media{Type: "image-external", Reference: "http://dummyimage.com/200x200/000/fff", Usage: "list"})

	product.Attributes = domain.Attributes(make(map[string]domain.Attribute))

	attributeValueList := make([]interface{}, 2)
	attributeValueList[0] = "departure"
	attributeValueList[1] = "arrival"

	product.Attributes = domain.Attributes{
		"brandCode":        domain.Attribute{RawValue: brands[rand.Intn(len(brands))]},
		"brandName":        domain.Attribute{RawValue: brands[rand.Intn(len(brands))]},
		"collectionOption": domain.Attribute{RawValue: attributeValueList},
	}

	product.RetailerCode = "om3CommonTestretailer"
	product.RetailerSku = "12345sku"
	product.CategoryPath = []string{"Testproducts", "Testproducts/Fake/Configurable"}
}

func (ps *ProductServiceAdapter) getPrice(defaultP float64, discounted float64) domain.PriceInfo {
	defaultP, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", defaultP), 64)
	discounted, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", discounted), 64)

	var price domain.PriceInfo
	price.Currency = "EUR"

	price.Default = defaultP
	if discounted != 0 {
		price.Discounted = discounted
		price.DiscountText = "Super test campaign"
		price.IsDiscounted = true
	}
	price.ActiveBase = 1
	price.ActiveBaseAmount = 10
	price.ActiveBaseUnit = "ml"
	return price
}
*/

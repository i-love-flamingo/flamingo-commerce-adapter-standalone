package productRepository

import (
	"errors"
	"fmt"
	"strings"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/csv"
	"flamingo.me/flamingo-commerce/product/domain"
	"flamingo.me/flamingo/framework/flamingo"
)

type (
	InMemoryProductRepositoryFactory struct {
		Logger flamingo.Logger `inject:""`
	}
)

func (f *InMemoryProductRepositoryFactory) BuildFromProductCSV(csvFile string, locale string) (*InMemoryProductRepository, error) {
	rows, err := csv.ReadProductCSV(csvFile)
	if err != nil {
		return nil, err
	}
	//todo use Dingo provider
	newRepo := InMemoryProductRepository{}

	for rowK, row := range rows {
		if row["productType"] == "simple" {
			product, err := f.buildSimpleProduct(row, locale)
			if err != nil {
				f.Logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
			}
			newRepo.add(product)
		}
	}
	return &newRepo, nil
}

func (f *InMemoryProductRepositoryFactory) buildSimpleProduct(row map[string]string, locale string) (*domain.SimpleProduct, error) {
	for _, requiredAttribute := range []string{"marketplaceCode", "retailerCode", "title-" + locale, "metaKeywords-" + locale, "shortDescription-" + locale, "description-" + locale} {
		if _, ok := row[requiredAttribute]; !ok {
			return nil, errors.New("marketplaceCode" + " Is missing (required attribute)")
		}
	}

	simple := domain.SimpleProduct{
		Identifier: row["marketplaceCode"],
		BasicProductData: domain.BasicProductData{
			MarketPlaceCode:  row["marketplaceCode"],
			RetailerCode:     row["retailerCode"],
			CategoryCodes:    strings.Split(row["categories"], ","),
			Title:            row["title-"+locale],
			ShortDescription: row["title-"+locale],
			Description:      row["ddddd-"+locale],
			RetailerName:     row["retailerCode"],
			Media:            f.getMedia(row, locale),
			Keywords:         strings.Split("metaKeywords-"+locale, ","),
		},
	}
	return &simple, nil
}

func (f *InMemoryProductRepositoryFactory) getMedia(row map[string]string, locale string) []domain.Media {
	var medias []domain.Media
	return medias
}

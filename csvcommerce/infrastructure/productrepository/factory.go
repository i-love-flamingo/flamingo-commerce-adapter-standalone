package productrepository

import (
	"errors"
	"fmt"
	"strings"

	priceDomain "flamingo.me/flamingo-commerce/v3/price/domain"

	"strconv"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvcommerce/infrastructure/csv"
	inMemoryProductSearchInfrastructure "flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
)

type (
	// InMemoryProductRepositoryFactory returns a Product Repository Type which is held in memory
	InMemoryProductRepositoryFactory struct {
		logger flamingo.Logger
	}
)

// Inject method to inject dependencies
func (f *InMemoryProductRepositoryFactory) Inject(logger flamingo.Logger) {
	f.logger = logger
}

// BuildFromProductCSV reads Products from a CSV File and returns a Product Repository of the In Memory Type
func (f *InMemoryProductRepositoryFactory) BuildFromProductCSV(csvFile string, locale string, currency string) (*inMemoryProductSearchInfrastructure.InMemoryProductRepository, error) {
	rows, err := csv.ReadProductCSV(csvFile)
	if err != nil {
		return nil, err
	}

	newRepo := inMemoryProductSearchInfrastructure.InMemoryProductRepository{}

	for rowK, row := range rows {
		if row["productType"] == "simple" {
			product, err := f.buildSimpleProduct(row, locale, currency)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
				continue
			}

			newRepo.Add(*product)
		}
	}

	for rowK, row := range rows {
		if row["productType"] == "configurable" {
			product, err := f.buildConfigurableProduct(newRepo, row, locale, currency)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
				continue
			}

			newRepo.Add(*product)
		}
	}

	return &newRepo, nil
}

// buildConfigurableProduct creates Products of the Configurable Type from CSV Rows
func (f *InMemoryProductRepositoryFactory) buildConfigurableProduct(repo inMemoryProductSearchInfrastructure.InMemoryProductRepository, row map[string]string, locale string, currency string) (*domain.ConfigurableProduct, error) {
	err := f.validateRow(row, locale, currency, []string{"variantVariationAttributes", "CONFIGURABLE-products"})
	if err != nil {
		return nil, err
	}
	configurable := domain.ConfigurableProduct{
		Identifier:       f.getIdentifier(row),
		BasicProductData: f.getBasicProductData(row, locale),
	}

	variantCodes := splitTrimmed(row["CONFIGURABLE-products"])
	if len(variantCodes) == 0 {
		return nil, errors.New("No  CONFIGURABLE-products entries in CSV found")
	}

	for _, vcode := range variantCodes {
		variantProduct, err := repo.FindByMarketplaceCode(vcode)
		if err != nil {
			return nil, err
		}
		configurable.Variants = append(configurable.Variants,
			domain.Variant{
				BasicProductData: variantProduct.BaseData(),
				Saleable:         variantProduct.SaleableData(),
			})
	}

	configurable.VariantVariationAttributes = splitTrimmed(row["variantVariationAttributes"])

	return &configurable, nil
}

// splitTrimmed splits strings by comma and returns a slice of pre-trimmed strings
func splitTrimmed(value string) []string {
	result := strings.Split(value, ",")
	for k, v := range result {
		result[k] = strings.TrimSpace(v)
	}

	return result
}

// validateRow ensures CSV Rows have the correct columns
func (f *InMemoryProductRepositoryFactory) validateRow(row map[string]string, locale string, currency string, additionalRequiredCols []string) error {
	additionalRequiredCols = append(additionalRequiredCols, []string{"marketplaceCode", "retailerCode", "title-" + locale, "metaKeywords-" + locale, "shortDescription-" + locale, "description-" + locale, "price-" + currency}...)
	for _, requiredAttribute := range additionalRequiredCols {
		if _, ok := row[requiredAttribute]; !ok {
			return fmt.Errorf("required attribute %q is missing", requiredAttribute)
		}
	}

	return nil
}

// getBasicProductData reads a CSV row and returns Basic Product Data Structs
func (f *InMemoryProductRepositoryFactory) getBasicProductData(row map[string]string, locale string) domain.BasicProductData {
	attributes := make(map[string]domain.Attribute)

	for key, data := range row {
		attributes[key] = domain.Attribute{
			Code:     key,
			Label:    key,
			RawValue: data,
		}
	}

	return domain.BasicProductData{
		MarketPlaceCode:  row["marketplaceCode"],
		RetailerCode:     row["retailerCode"],
		CategoryCodes:    strings.Split(row["categories"], ","),
		Title:            row["title-"+locale],
		ShortDescription: row["shortDescription-"+locale],
		Description:      row["description-"+locale],
		RetailerName:     row["retailerCode"],
		Media:            f.getMedia(row, locale),
		Keywords:         strings.Split("metaKeywords-"+locale, ","),
		Attributes:       attributes,
	}
}

// getIdentifier returns only the Product Identifier (aka marketPlaceCode) from a map of strings (previously CSV Row)
func (f *InMemoryProductRepositoryFactory) getIdentifier(row map[string]string) string {
	return row["marketplaceCode"]
}

// buildSimpleProduct builds a Product of the Simple Type from a map of strings (previously a CSV Row)
func (f *InMemoryProductRepositoryFactory) buildSimpleProduct(row map[string]string, locale string, currency string) (*domain.SimpleProduct, error) {
	err := f.validateRow(row, locale, currency, nil)
	if err != nil {
		return nil, err
	}

	price, _ := strconv.ParseFloat(row["price-"+currency], 64)

	simple := domain.SimpleProduct{
		Identifier:       f.getIdentifier(row),
		BasicProductData: f.getBasicProductData(row, locale),
		Saleable: domain.Saleable{
			ActivePrice: domain.PriceInfo{
				Default: priceDomain.NewFromFloat(price, currency),
			},
		},
	}

	simple.Teaser = domain.TeaserData{
		TeaserPrice: simple.Saleable.ActivePrice,
	}

	return &simple, nil
}

// getMedia gets the Product Images from a map of strings (previously a CSV Row)
func (f *InMemoryProductRepositoryFactory) getMedia(row map[string]string, locale string) []domain.Media {
	var medias []domain.Media
	if v, ok := row["listImage"]; ok {
		medias = append(medias, domain.Media{
			Type:      "csvCommerceReference",
			Reference: v,
			Usage:     domain.MediaUsageList,
		})
	}
	if v, ok := row["thumbnailImage"]; ok {
		medias = append(medias, domain.Media{
			Type:      "csvCommerceReference",
			Reference: v,
			Usage:     domain.MediaUsageThumbnail,
		})
	}
	for _, dk := range []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10"} {
		if v, ok := row["detailImage"+dk]; ok {
			medias = append(medias, domain.Media{
				Type:      "csvCommerceReference",
				Reference: v,
				Usage:     domain.MediaUsageDetail,
			})
		}
	}

	return medias
}

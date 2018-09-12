package productRepository

import (
	"errors"
	"fmt"
	"strings"

	"strconv"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvCommerce/infrastructure/csv"
	inMemoryProductSearchInfrastructure "flamingo.me/flamingo-commerce-adapter-standalone/inMemoryProductSearch/infrastructure"
	"flamingo.me/flamingo-commerce/product/domain"
	"flamingo.me/flamingo/framework/flamingo"
)

type (
	InMemoryProductRepositoryFactory struct {
		Logger flamingo.Logger `inject:""`
	}
	InMemoryProductRepositoryProvider struct {
		InMemoryProductRepositoryFactory *InMemoryProductRepositoryFactory `inject:""`
		Logger                           flamingo.Logger                   `inject:""`
		Locale                           string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.locale"`
		Currency                         string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.currency"`
		ProductCSVPath                   string                            `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
	}
)

//TODO - use map with lock (sync map) https://github.com/golang/go/blob/master/src/sync/map.go
var buildedRepositoryByLocale = make(map[string]*inMemoryProductSearchInfrastructure.InMemoryProductRepository)

func (f *InMemoryProductRepositoryFactory) BuildFromProductCSV(csvFile string, locale string, currency string) (*inMemoryProductSearchInfrastructure.InMemoryProductRepository, error) {
	rows, err := csv.ReadProductCSV(csvFile)
	if err != nil {
		return nil, err
	}
	//todo use Dingo provider
	newRepo := inMemoryProductSearchInfrastructure.InMemoryProductRepository{}

	for rowK, row := range rows {
		if row["productType"] == "simple" {
			product, err := f.buildSimpleProduct(row, locale, currency)
			if err != nil {
				f.Logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
			}
			newRepo.Add(*product)
		}
	}
	for rowK, row := range rows {
		if row["productType"] == "configurable" {
			product, err := f.buildConfigurableProduct(newRepo, row, locale, currency)
			if err != nil {
				f.Logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
			}
			newRepo.Add(*product)
		}
	}

	return &newRepo, nil
}

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

func splitTrimmed(value string) []string {
	result := strings.Split(value, ",")
	for k, v := range result {
		result[k] = strings.TrimSpace(v)
	}
	return result
}

func (f *InMemoryProductRepositoryFactory) validateRow(row map[string]string, locale string, currency string, additionalRequiredCols []string) error {
	additionalRequiredCols = append(additionalRequiredCols, []string{"marketplaceCode", "retailerCode", "title-" + locale, "metaKeywords-" + locale, "shortDescription-" + locale, "description-" + locale, "price-" + currency}...)
	for _, requiredAttribute := range additionalRequiredCols {
		if _, ok := row[requiredAttribute]; !ok {
			return errors.New("\"" + requiredAttribute + "\"" + " Is missing (required attribute)")
		}
	}
	return nil
}

func (f *InMemoryProductRepositoryFactory) getBasicProductData(row map[string]string, locale string) domain.BasicProductData {
	attributes := make(map[string]domain.Attribute)

	for key, data := range row {
		attributes[key] = domain.Attribute{
			Code: key,
			Label: key,
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

func (f *InMemoryProductRepositoryFactory) getIdentifier(row map[string]string) string {
	return row["marketplaceCode"]
}

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
				Default:  price,
				Currency: currency,
			},
		},
	}
	simple.Teaser = domain.TeaserData{
		TeaserPrice: simple.Saleable.ActivePrice,
	}
	return &simple, nil
}

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

func (p *InMemoryProductRepositoryProvider) GetForCurrentLocale() (*inMemoryProductSearchInfrastructure.InMemoryProductRepository, error) {
	locale := p.Locale
	if v, ok := buildedRepositoryByLocale[locale]; ok {
		return v, nil
	}
	p.Logger.Info("Build InMemoryProductRepository for locale " + locale + " .....")
	rep, err := p.InMemoryProductRepositoryFactory.BuildFromProductCSV(p.ProductCSVPath, locale, p.Currency)
	if err != nil {
		return nil, err
	}
	buildedRepositoryByLocale[locale] = rep
	return buildedRepositoryByLocale[locale], nil
}

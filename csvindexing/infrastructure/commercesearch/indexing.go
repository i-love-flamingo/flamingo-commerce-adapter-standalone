package commercesearch

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	categorydomain "flamingo.me/flamingo-commerce/v3/category/domain"
	priceDomain "flamingo.me/flamingo-commerce/v3/price/domain"
	"flamingo.me/flamingo-commerce/v3/product/domain"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"

	commerceSearchDomain "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/infrastructure/csv"
)

type (
	// IndexUpdater implements indexing based on CSV file
	IndexUpdater struct {
		logger                   flamingo.Logger
		productCsvFile           string
		productCsvDelimiter      rune
		productAttributesToSplit map[string]struct{}
		categoryCsvFile          string
		categoryCsvDelimiter     rune
		categoryTreeBuilder      *commerceSearchDomain.CategoryTreeBuilder
		locale                   string
		currency                 string
	}
)

var (
	_ commerceSearchDomain.IndexUpdater = &IndexUpdater{}
)

// Inject method to inject dependencies
func (f *IndexUpdater) Inject(logger flamingo.Logger, categoryTreeBuilder *commerceSearchDomain.CategoryTreeBuilder,
	config *struct {
		ProductCsvFile           string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.path"`
		ProductCsvDelimiter      string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.delimiter"`
		ProductAttributesToSplit config.Slice `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.attributesToSplit"`
		CategoryCsvFile          string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.path,optional"`
		CategoryCsvDelimiter     string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.delimiter,optional"`
		Locale                   string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.locale"`
		Currency                 string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.currency"`
	}) {
	f.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone.csvindexing").WithField(flamingo.LogKeyCategory, "IndexUpdater")
	f.categoryTreeBuilder = categoryTreeBuilder
	if config != nil {
		f.productCsvFile = config.ProductCsvFile
		if config.ProductCsvDelimiter != "" {
			f.productCsvDelimiter = []rune(config.ProductCsvDelimiter)[0]
		}
		f.categoryCsvFile = config.CategoryCsvFile
		if config.CategoryCsvDelimiter != "" {
			f.categoryCsvDelimiter = []rune(config.CategoryCsvDelimiter)[0]
		}

		var toSplit []string
		err := config.ProductAttributesToSplit.MapInto(&toSplit)
		if err != nil {
			panic(err)
		}

		f.productAttributesToSplit = make(map[string]struct{})
		for _, attribute := range toSplit {
			f.productAttributesToSplit[attribute] = struct{}{}
		}

		f.locale = config.Locale
		f.currency = config.Currency
	}
}

// Index starts index process
func (f *IndexUpdater) Index(ctx context.Context, indexer *commerceSearchDomain.Indexer) error {
	f.logger.Info(fmt.Sprintf("Start loading CSV file: %v  with locale: %v and currency %v", f.productCsvFile, f.locale, f.currency))

	var err error
	var tree categorydomain.Tree
	// read category tree
	if f.categoryCsvFile != "" {
		catrows, err := csv.ReadCSV(f.categoryCsvFile, csv.DelimiterOption(f.productCsvDelimiter))
		if err != nil {
			return errors.New(err.Error() + " / File: " + f.categoryCsvFile)
		}
		for _, row := range catrows {
			f.categoryTreeBuilder.AddCategoryData(row["code"], row["label-"+f.locale], row["parent"])
		}
		tree, err = f.categoryTreeBuilder.BuildTree()
		// fmt.Printf("\n %#v",tree)
		// printTree(tree,"")
		if err != nil {
			return err
		}
	}

	// Index products
	rows, err := csv.ReadCSV(f.productCsvFile, csv.DelimiterOption(f.categoryCsvDelimiter))
	if err != nil {
		return errors.New(err.Error() + " / File: " + f.productCsvFile)
	}
	for rowK, row := range rows {
		if row["productType"] == "simple" {
			product, err := f.buildSimpleProduct(row, f.locale, f.currency, tree)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
				continue
			}

			err = indexer.UpdateProductAndCategory(ctx, *product)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error adding row %v (%v)", rowK, err))
			}
		}
	}

	for rowK, row := range rows {
		if row["productType"] == "configurable" {
			product, err := f.buildConfigurableProduct(ctx, indexer, row, f.locale, f.currency, tree)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error mapping row %v (%v)", rowK, err))
				continue
			}

			err = indexer.UpdateProductAndCategory(ctx, *product)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Error adding row %v (%v)", rowK, err))
			}
		}
	}
	return nil
}

// buildConfigurableProduct creates Products of the Configurable Type from CSV Rows
func (f *IndexUpdater) buildConfigurableProduct(ctx context.Context, indexer *commerceSearchDomain.Indexer, row map[string]string, locale string, currency string, tree categorydomain.Tree) (*domain.ConfigurableProduct, error) {
	err := f.validateRow(row, locale, currency, []string{"variantVariationAttributes", "CONFIGURABLE-products"})
	if err != nil {
		return nil, err
	}
	configurable := domain.ConfigurableProduct{
		Identifier:       f.getIdentifier(row),
		BasicProductData: f.getBasicProductData(row, locale, tree),
	}

	variantCodes := splitTrimmed(row["CONFIGURABLE-products"])
	if len(variantCodes) == 0 {
		return nil, errors.New("No  CONFIGURABLE-products entries in CSV found")
	}

	for _, vcode := range variantCodes {
		variantProduct, err := indexer.ProductRepository().FindByMarketplaceCode(ctx, vcode)
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
func (f *IndexUpdater) validateRow(row map[string]string, locale string, currency string, additionalRequiredCols []string) error {
	additionalRequiredCols = append(additionalRequiredCols,
		[]string{
			"marketplaceCode",
			"retailerCode",
			"title-" + locale,
			"metaKeywords-" + locale,
			"shortDescription-" + locale,
			"description-" + locale,
			"price-" + currency,
		}...)
	for _, requiredAttribute := range additionalRequiredCols {
		if _, ok := row[requiredAttribute]; !ok {
			return fmt.Errorf("required attribute %q is missing", requiredAttribute)
		}
	}

	return nil
}

// getBasicProductData reads a CSV row and returns Basic Product Data Structs
func (f *IndexUpdater) getBasicProductData(row map[string]string, locale string, tree categorydomain.Tree) domain.BasicProductData {
	attributes := make(map[string]domain.Attribute)

	for key, data := range row {

		// skip other locales
		parts := strings.Split(key, "-")
		if len(parts) > 1 {
			l := parts[len(parts)-1]
			if l != "" && l != locale {
				continue
			}
		}

		key = strings.TrimSuffix(key, "-"+locale)

		attributes[key] = domain.Attribute{
			Code:      key,
			CodeLabel: key,
			Label:     data,
			RawValue: func() interface{} {
				if _, found := f.productAttributesToSplit[key]; !found {
					return data
				}

				var split []interface{}
				for _, s := range strings.Split(data, ",") {
					split = append(split, s)
				}
				return split
			}(),
		}
	}

	var categories []domain.CategoryTeaser
	for _, categoryCode := range strings.Split(row["categories"], ",") {
		categoryTeaser := commerceSearchDomain.CategoryTreeToCategoryTeaser(categoryCode, tree)
		if categoryTeaser == nil {
			f.logger.Error(errors.New("categoryCode " + categoryCode + " not found in tree"))
		}

		if categoryTeaser != nil {
			categories = append(categories, *categoryTeaser)
		}
	}

	stockLevel := domain.StockLevelInStock
	switch row["stockLevel"] {
	case domain.StockLevelInStock,
		domain.StockLevelLowStock,
		domain.StockLevelOutOfStock:
		stockLevel = row["stockLevel"]
	}

	basicProductData := domain.BasicProductData{
		MarketPlaceCode:  row["marketplaceCode"],
		RetailerCode:     row["retailerCode"],
		Categories:       categories,
		Title:            row["title-"+locale],
		ShortDescription: row["shortDescription-"+locale],
		Description:      row["description-"+locale],
		RetailerName:     row["retailerName"],
		Media:            f.getMedia(row, locale),
		Keywords:         strings.Split("metaKeywords-"+locale, ","),
		Attributes:       attributes,
		StockLevel:       stockLevel,
	}
	if len(categories) > 0 {
		basicProductData.MainCategory = categories[0]
	}
	return basicProductData
}

// getIdentifier returns only the Product Identifier (aka marketPlaceCode) from a map of strings (previously CSV Row)
func (f *IndexUpdater) getIdentifier(row map[string]string) string {
	return row["marketplaceCode"]
}

// buildSimpleProduct builds a Product of the Simple Type from a map of strings (previously a CSV Row)
func (f *IndexUpdater) buildSimpleProduct(row map[string]string, locale string, currency string, tree categorydomain.Tree) (*domain.SimpleProduct, error) {
	err := f.validateRow(row, locale, currency, nil)
	if err != nil {
		return nil, err
	}

	price, _ := strconv.ParseFloat(row["price-"+currency], 64)
	specialPrice, specialPriceErr := strconv.ParseFloat(row["specialPrice-"+currency], 64)
	hasSpecialPrice := false
	if specialPriceErr == nil && specialPrice != price {
		hasSpecialPrice = true
	}

	isSaleable := true
	if _, ok := row["saleable"]; ok {
		isSaleable, _ = strconv.ParseBool(row["saleable"])
	}

	saleableFrom := time.Time{}
	if from, ok := row["saleableFromDate"]; ok {
		saleableFrom, _ = time.Parse(time.RFC3339, from)
	}

	saleableTo := time.Time{}
	if from, ok := row["saleableToDate"]; ok {
		saleableTo, _ = time.Parse(time.RFC3339, from)
	}

	simple := domain.SimpleProduct{
		Identifier:       f.getIdentifier(row),
		BasicProductData: f.getBasicProductData(row, locale, tree),
		Saleable: domain.Saleable{
			IsSaleable:   isSaleable,
			SaleableFrom: saleableFrom,
			SaleableTo:   saleableTo,
			ActivePrice: domain.PriceInfo{
				Default:      priceDomain.NewFromFloat(price, currency).GetPayable(),
				IsDiscounted: hasSpecialPrice,
				Discounted:   priceDomain.NewFromFloat(specialPrice, currency).GetPayable(),
			},
		},
	}

	simple.Teaser = domain.TeaserData{
		ShortTitle:       simple.BasicProductData.Title,
		ShortDescription: simple.BasicProductData.ShortDescription,
		TeaserPrice:      simple.Saleable.ActivePrice,
		Media:            simple.BaseData().Media,
		MarketPlaceCode:  simple.BasicProductData.MarketPlaceCode,
	}

	return &simple, nil
}

// getMedia gets the Product Images from a map of strings (previously a CSV Row)
func (f *IndexUpdater) getMedia(row map[string]string, locale string) []domain.Media {
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

func printTree(tree categorydomain.Tree, indend string) {
	fmt.Printf("\n %v > %v", indend, tree.Code())
	for _, s := range tree.SubTrees() {
		printTree(s, indend+"   ")
	}
}

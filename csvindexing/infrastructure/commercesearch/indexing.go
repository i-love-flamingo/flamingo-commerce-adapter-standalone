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
	productDomain "flamingo.me/flamingo-commerce/v3/product/domain"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"

	commerceSearchDomain "flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/infrastructure/csv"
)

type (
	// IndexUpdater implements indexing based on CSV file
	IndexUpdater struct {
		logger                   flamingo.Logger
		productRowPreprocessors  []domain.ProductRowPreprocessor
		categoryRowPreprocessors []domain.CategoryRowPreprocessor
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
func (u *IndexUpdater) Inject(
	logger flamingo.Logger,
	categoryTreeBuilder *commerceSearchDomain.CategoryTreeBuilder,
	productRowPreprocessors []domain.ProductRowPreprocessor,
	categoryRowPreprocessors []domain.CategoryRowPreprocessor,
	config *struct {
		ProductCsvFile           string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.path"`
		ProductCsvDelimiter      string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.delimiter"`
		ProductAttributesToSplit config.Slice `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.attributesToSplit"`
		CategoryCsvFile          string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.path,optional"`
		CategoryCsvDelimiter     string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.delimiter,optional"`
		Locale                   string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.locale"`
		Currency                 string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.currency"`
	}) *IndexUpdater {
	u.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone.csvindexing").WithField(flamingo.LogKeyCategory, "IndexUpdater")
	u.categoryTreeBuilder = categoryTreeBuilder
	u.productRowPreprocessors = productRowPreprocessors
	u.categoryRowPreprocessors = categoryRowPreprocessors
	if config != nil {
		u.productCsvFile = config.ProductCsvFile
		if config.ProductCsvDelimiter != "" {
			u.productCsvDelimiter = []rune(config.ProductCsvDelimiter)[0]
		}
		u.categoryCsvFile = config.CategoryCsvFile
		if config.CategoryCsvDelimiter != "" {
			u.categoryCsvDelimiter = []rune(config.CategoryCsvDelimiter)[0]
		}

		var toSplit []string
		err := config.ProductAttributesToSplit.MapInto(&toSplit)
		if err != nil {
			panic(err)
		}

		u.productAttributesToSplit = make(map[string]struct{})
		for _, attribute := range toSplit {
			u.productAttributesToSplit[attribute] = struct{}{}
		}

		u.locale = config.Locale
		u.currency = config.Currency
	}

	return u
}

// Index starts index process
func (u *IndexUpdater) Index(ctx context.Context, indexer *commerceSearchDomain.Indexer) error {
	u.logger.Info(fmt.Sprintf("Start loading CSV file: %v  with locale: %v and currency %v", u.productCsvFile, u.locale, u.currency))

	var err error
	var tree categorydomain.Tree
	// read category tree
	if u.categoryCsvFile != "" {
		catRows, err := csv.ReadCSV(u.categoryCsvFile, csv.DelimiterOption(u.categoryCsvDelimiter))
		if err != nil {
			return errors.New(err.Error() + " / File: " + u.categoryCsvFile)
		}
		for rowK, row := range catRows {
			for _, preprocessor := range u.categoryRowPreprocessors {
				row, err = preprocessor.Preprocess(
					row,
					domain.CategoryRowPreprocessOptions{
						Locale: u.locale,
					},
				)
				if err != nil {
					u.logger.Error(fmt.Sprintf("Preprocessing: %s / Row: %d, File: %s", err, rowK, u.categoryCsvFile))
				}
			}
			err = u.validateCategoryRow(row)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Validating: %s / Row: %d, File: %s", err, rowK, u.categoryCsvFile))
				continue
			}
			u.categoryTreeBuilder.AddCategoryData(row["code"], row["label-"+u.locale], row["parent"])
		}
		tree, err = u.categoryTreeBuilder.BuildTree()
		if err != nil {
			return err
		}
	}

	// Index products
	rows, err := csv.ReadCSV(u.productCsvFile, csv.DelimiterOption(u.productCsvDelimiter))
	if err != nil {
		return errors.New(err.Error() + " / File: " + u.productCsvFile)
	}
	for rowK, row := range rows {
		for _, preprocessor := range u.productRowPreprocessors {
			row, err = preprocessor.Preprocess(
				row,
				domain.ProductRowPreprocessOptions{
					Locale:   u.locale,
					Currency: u.currency,
				},
			)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Preprocessing: %s / Row: %d, File: %s", err, rowK, u.productCsvFile))
			}
			rows[rowK] = row
		}
	}
	for rowK, row := range rows {
		if row["productType"] == "simple" {
			product, err := u.buildSimpleProduct(row, tree)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Mapping: %s / Row: %d, File: %s", err, rowK, u.productCsvFile))
				continue
			}

			err = indexer.UpdateProductAndCategory(ctx, *product)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Adding: %s / Row: %d, File: %s", err, rowK, u.productCsvFile))
			}
		}
	}

	for rowK, row := range rows {
		if row["productType"] == "configurable" {
			product, err := u.buildConfigurableProduct(ctx, indexer, row, tree)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Mapping: %s / Row: %d, File: %s", err, rowK, u.productCsvFile))
				continue
			}

			err = indexer.UpdateProductAndCategory(ctx, *product)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Adding: %s / Row: %d, File: %s", err, rowK, u.productCsvFile))
			}
		}
	}
	return nil
}

// buildConfigurableProduct creates Products of the Configurable Type from CSV Rows
func (u *IndexUpdater) buildConfigurableProduct(ctx context.Context, indexer *commerceSearchDomain.Indexer, row map[string]string, tree categorydomain.Tree) (*productDomain.ConfigurableProduct, error) {
	err := u.validateRow(row, []string{"variantVariationAttributes", "CONFIGURABLE-products"})
	if err != nil {
		return nil, err
	}
	configurable := productDomain.ConfigurableProduct{
		Identifier:       u.getIdentifier(row),
		BasicProductData: u.getBasicProductData(row, tree),
	}

	variantCodes := splitTrimmed(row["CONFIGURABLE-products"])
	if len(variantCodes) == 0 {
		return nil, errors.New("no CONFIGURABLE-products entries in CSV found")
	}

	for _, vcode := range variantCodes {
		variantProduct, err := indexer.ProductRepository().FindByMarketplaceCode(ctx, vcode)
		if err != nil {
			return nil, err
		}
		configurable.Variants = append(configurable.Variants,
			productDomain.Variant{
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
func (u *IndexUpdater) validateRow(row map[string]string, additionalRequiredCols []string) error {
	additionalRequiredCols = append(additionalRequiredCols,
		[]string{
			"marketplaceCode",
			"retailerCode",
			"title-" + u.locale,
			"description-" + u.locale,
		}...)
	for _, requiredAttribute := range additionalRequiredCols {
		if val, ok := row[requiredAttribute]; !ok || val == "" {
			return fmt.Errorf("required column %q is missing", requiredAttribute)
		}
	}

	return nil
}

func (u *IndexUpdater) validateCategoryRow(row csv.RowDto) error {
	requiredCols := []string{
		"code",
		"parent",
		"label-" + u.locale,
	}
	for _, requiredAttribute := range requiredCols {
		if val, ok := row[requiredAttribute]; !ok || val == "" {
			return fmt.Errorf("required column %q is missing", requiredAttribute)
		}
	}

	return nil
}

// getBasicProductData reads a CSV row and returns Basic Product Data Structs
func (u *IndexUpdater) getBasicProductData(row map[string]string, tree categorydomain.Tree) productDomain.BasicProductData {
	attributes := make(map[string]productDomain.Attribute)

	for key, data := range row {
		// skip empty fields
		if data == "" {
			continue
		}

		// skip other locales
		parts := strings.Split(key, "-")
		if len(parts) > 1 {
			l := parts[len(parts)-1]
			if l != "" && l != u.locale {
				continue
			}
		}

		key = strings.TrimSuffix(key, "-"+u.locale)

		attributes[key] = productDomain.Attribute{
			Code:      key,
			CodeLabel: key,
			Label:     data,
			RawValue: func() interface{} {
				if _, found := u.productAttributesToSplit[key]; !found {
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

	var categories []productDomain.CategoryTeaser
	for _, categoryCode := range strings.Split(row["categories"], ",") {
		categoryTeaser := commerceSearchDomain.CategoryTreeToCategoryTeaser(categoryCode, tree)
		if categoryTeaser == nil {
			u.logger.Error(errors.New("categoryCode " + categoryCode + " not found in tree"))
		}

		if categoryTeaser != nil {
			categories = append(categories, *categoryTeaser)
		}
	}

	stockLevel := productDomain.StockLevelInStock
	switch row["stockLevel"] {
	case productDomain.StockLevelInStock,
		productDomain.StockLevelLowStock,
		productDomain.StockLevelOutOfStock:
		stockLevel = row["stockLevel"]
	}

	basicProductData := productDomain.BasicProductData{
		MarketPlaceCode:  row["marketplaceCode"],
		RetailerCode:     row["retailerCode"],
		Categories:       categories,
		Title:            row["title-"+u.locale],
		ShortDescription: row["shortDescription-"+u.locale],
		Description:      row["description-"+u.locale],
		RetailerName:     row["retailerName"],
		Media:            u.getMedia(row),
		Keywords:         strings.Split(row["metaKeywords-"+u.locale], ","),
		Attributes:       attributes,
		StockLevel:       stockLevel,
	}
	if len(categories) > 0 {
		basicProductData.MainCategory = categories[0]
	}
	return basicProductData
}

// getIdentifier returns only the Product Identifier (aka marketPlaceCode) from a map of strings (previously CSV Row)
func (u *IndexUpdater) getIdentifier(row map[string]string) string {
	return row["marketplaceCode"]
}

// buildSimpleProduct builds a Product of the Simple Type from a map of strings (previously a CSV Row)
func (u *IndexUpdater) buildSimpleProduct(row map[string]string, tree categorydomain.Tree) (*productDomain.SimpleProduct, error) {
	err := u.validateRow(row, []string{"price-" + u.currency})
	if err != nil {
		return nil, err
	}

	price, _ := strconv.ParseFloat(row["price-"+u.currency], 64)
	specialPrice, specialPriceErr := strconv.ParseFloat(row["specialPrice-"+u.currency], 64)
	hasSpecialPrice := false
	if specialPriceErr == nil && specialPrice != price {
		hasSpecialPrice = true
	}

	isSaleable := true
	if _, ok := row["saleable"]; ok {
		isSaleable, _ = strconv.ParseBool(row["saleable"])
	}

	saleableFrom := time.Time{}
	if from, ok := row["saleableFromDate"]; ok && from != "" {
		saleableFrom, _ = time.Parse(time.RFC3339, from)
	}

	saleableTo := time.Time{}
	if from, ok := row["saleableToDate"]; ok && from != "" {
		saleableTo, _ = time.Parse(time.RFC3339, from)
	}

	simple := productDomain.SimpleProduct{
		Identifier:       u.getIdentifier(row),
		BasicProductData: u.getBasicProductData(row, tree),
		Saleable: productDomain.Saleable{
			IsSaleable:   isSaleable,
			SaleableFrom: saleableFrom,
			SaleableTo:   saleableTo,
			ActivePrice: productDomain.PriceInfo{
				Default:      priceDomain.NewFromFloat(price, u.currency).GetPayable(),
				IsDiscounted: hasSpecialPrice,
				Discounted:   priceDomain.NewFromFloat(specialPrice, u.currency).GetPayable(),
			},
		},
	}

	simple.Teaser = productDomain.TeaserData{
		ShortTitle:       simple.BasicProductData.Title,
		ShortDescription: simple.BasicProductData.ShortDescription,
		TeaserPrice:      simple.Saleable.ActivePrice,
		Media:            simple.BaseData().Media,
		MarketPlaceCode:  simple.BasicProductData.MarketPlaceCode,
	}

	return &simple, nil
}

// getMedia gets the Product Images from a map of strings (previously a CSV Row)
func (u *IndexUpdater) getMedia(row map[string]string) []productDomain.Media {
	var medias []productDomain.Media
	if v, ok := row["listImage"]; ok && v != "" {
		medias = append(medias, productDomain.Media{
			Type:      "csvCommerceReference",
			Reference: v,
			Usage:     productDomain.MediaUsageList,
		})
	}
	if v, ok := row["thumbnailImage"]; ok && v != "" {
		medias = append(medias, productDomain.Media{
			Type:      "csvCommerceReference",
			Reference: v,
			Usage:     productDomain.MediaUsageThumbnail,
		})
	}
	for _, dk := range []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10"} {
		if v, ok := row["detailImage"+dk]; ok && v != "" {
			medias = append(medias, productDomain.Media{
				Type:      "csvCommerceReference",
				Reference: v,
				Usage:     productDomain.MediaUsageDetail,
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

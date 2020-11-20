package commercesearch_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"testing"

	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/domain"
	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch/infrastructure/productsearch"
	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing/infrastructure/commercesearch"
)

// Expect 35 minutes for a run
// go test -bench . -benchtime=60s -timeout=60m

// prepareBigCSV takes the header of the products.csv and creates a CSV with n products which are a copy from
// the last row of products.csv with just an incremented SKU
func prepareBigCSV(b *testing.B, n int) {
	in, _ := os.Open("../../testdata/products.csv")
	defer in.Close()

	reader := csv.NewReader(in)
	rows, err := reader.ReadAll()
	require.NoError(b, err, "base CSV file not readable")
	h := rows[0]
	r := rows[len(rows)-1]

	targetFileName := fmt.Sprintf("../../testdata/productsXL-%d.csv", n)
	file, err := os.Create(targetFileName)
	require.NoError(b, err, "target csv not writable")
	b.Cleanup(func() {
		os.Remove(targetFileName)
	})
	writer := csv.NewWriter(file)
	require.NoError(b, writer.Write(h))
	require.NoError(b, writer.Write(r))

	for i := 0; i < n; i++ {
		sku, _ := strconv.Atoi(r[0])
		r[0] = strconv.Itoa(sku + 1)
		require.NoError(b, writer.Write(r))
	}

	writer.Flush()
	require.NoError(b, file.Close())
}

func BenchmarkIndexUpdater_IndexSmall(b *testing.B) {
	// ~30 products
	benchmarkIndexer(
		b,
		"../../testdata/products.csv",
		"../../testdata/categories.csv",
		"1000003",
	)
}

func BenchmarkIndexUpdater_IndexXL100(b *testing.B) {
	prepareBigCSV(b, 100)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-100.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL200(b *testing.B) {
	prepareBigCSV(b, 200)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-200.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL300(b *testing.B) {
	prepareBigCSV(b, 300)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-300.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL400(b *testing.B) {
	prepareBigCSV(b, 400)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-400.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL500(b *testing.B) {
	prepareBigCSV(b, 500)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-500.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL600(b *testing.B) {
	prepareBigCSV(b, 600)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-600.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL700(b *testing.B) {
	prepareBigCSV(b, 700)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-700.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL800(b *testing.B) {
	prepareBigCSV(b, 800)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-800.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL900(b *testing.B) {
	prepareBigCSV(b, 900)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-900.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL1000(b *testing.B) {
	prepareBigCSV(b, 1000)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-1000.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL5000(b *testing.B) {
	prepareBigCSV(b, 5000)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-5000.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func BenchmarkIndexUpdater_IndexXL10000(b *testing.B) {
	prepareBigCSV(b, 10000)
	benchmarkIndexer(
		b,
		"../../testdata/productsXL-10000.csv",
		"../../testdata/categories.csv",
		"1000010",
	)
}

func benchmarkIndexer(b *testing.B, productsCsv string, categoriesCsv string, checkMarketplaceCode string) {
	for n := 0; n < b.N; n++ {
		rep := new(productsearch.BleveRepository).Inject(
			new(flamingo.NullLogger),
			nil,
		)

		indexer := new(domain.Indexer).Inject(
			new(flamingo.NullLogger),
			rep,
			&struct {
				CategoryRepository domain.CategoryRepository `inject:",optional"`
			}{
				CategoryRepository: rep,
			},
		)
		require.NoError(b, indexer.PrepareIndex(context.Background()))

		indexUpdater := commercesearch.IndexUpdater{}

		indexUpdater.Inject(flamingo.NullLogger{},
			&domain.CategoryTreeBuilder{},
			&struct {
				ProductCsvFile           string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.path"`
				ProductCsvDelimiter      string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.file.delimiter"`
				ProductAttributesToSplit config.Slice `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.products.attributesToSplit"`
				CategoryCsvFile          string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.path,optional"`
				CategoryCsvDelimiter     string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.categories.file.delimiter,optional"`
				Locale                   string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.locale"`
				Currency                 string       `inject:"config:flamingoCommerceAdapterStandalone.csvindexing.currency"`
			}{
				Currency:        "GBP",
				Locale:          "en_GB",
				ProductCsvFile:  productsCsv,
				CategoryCsvFile: categoriesCsv,
			},
		)

		err2 := indexUpdater.Index(context.Background(), indexer)
		assert.NoError(b, err2)

		_, err := indexer.ProductRepository().FindByMarketplaceCode(context.Background(), checkMarketplaceCode)
		assert.NoError(b, err)
	}
}

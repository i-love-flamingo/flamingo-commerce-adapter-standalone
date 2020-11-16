package csvindexing_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/config"

	"flamingo.me/flamingo-commerce-adapter-standalone/csvindexing"
)

func TestModule_Configure(t *testing.T) {
	if err := config.TryModules(nil, new(csvindexing.ProductModule)); err != nil {
		t.Error(err)
	}
}

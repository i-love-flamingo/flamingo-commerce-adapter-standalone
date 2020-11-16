package commercesearch_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/config"

	"flamingo.me/flamingo-commerce-adapter-standalone/commercesearch"
)

func TestModule_Configure(t *testing.T) {
	if err := config.TryModules(nil, new(commercesearch.Module)); err != nil {
		t.Error(err)
	}
}

package emailplaceorder_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/config"

	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder"
)

func TestModule_Configure(t *testing.T) {
	if err := config.TryModules(config.Map{
		"core.auth.web.debugController":          false,
		"commerce.cart.placeOrderLogger.enabled": false,
	}, new(emailplaceorder.Module)); err != nil {
		t.Error(err)
	}
}

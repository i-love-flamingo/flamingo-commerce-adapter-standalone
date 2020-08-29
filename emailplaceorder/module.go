package emailplaceorder

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure"
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure/template"
	"flamingo.me/flamingo-commerce/v3/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	priceApp "flamingo.me/flamingo-commerce/v3/price/application"
)

type (
	// Module registers our profiler
	Module struct {
	}
)

// Configure module
func (m *Module) Configure(injector *dingo.Injector) {
	injector.Bind((*placeorder.Service)(nil)).To(infrastructure.PlaceOrderServiceAdapter{})
	injector.Bind(new(infrastructure.MailSender)).To(infrastructure.DefaultMailSender{})
	injector.Bind(new(infrastructure.MailTemplate)).To(template.Default{})
	injector.Bind(new(template.PriceFormat)).To(priceApp.Service{})
}

// CueConfig defines the cart module configuration
func (*Module) CueConfig() string {
	return `
flamingoCommerceAdapterStandalone: {
	emailplaceorder: {
		emailAddress: string | *""
		fromMail: string | *""
		fromName: string | *""
		credentials: {
			password: string | *""
			server:   string | *""
			port:     string | *""
			user:     string | *""
		}
	}
}`
}

// Depends on other modules
func (m *Module) Depends() []dingo.Module {
	return []dingo.Module{
		new(cart.Module),
	}
}

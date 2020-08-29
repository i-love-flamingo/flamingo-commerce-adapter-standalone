package template

import (
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure"
	cartDomain "flamingo.me/flamingo-commerce/v3/cart/domain/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	"fmt"
)

// AdminMail for store owner
func (d *Default) AdminMail(cart *cartDomain.Cart, payment *placeorder.Payment) (*infrastructure.Mail, error) {

	contentTemplate := "<h1>You received an order {{  .Cart.GetPaymentReference }}  for {{  .Cart.GetContactMail }} {{  priceFormat .Cart.GrandTotal }}</h1>"
	contentTemplate = contentTemplate +
		`
{{ if .Cart.BillingAddress }}
	{{ with  .Cart.BillingAddress }}
` + d.AddressTemplate() + `
	{{ end }}
{{ end }}

`
	contentTemplate = contentTemplate + d.BasketTemplate()

	template := d.WrapHTMLTemplate(contentTemplate, "")
	result, err := d.GenerateTemplate(cart, payment, template)
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}
	return &infrastructure.Mail{
		HTML:    result,
		Plain:   fmt.Sprintf("You received an order for %v / Total: %v", cart.GetContactMail(), d.priceFormat.FormatPrice(cart.GrandTotal())),
		Subject: fmt.Sprintf("You received an order - %v", cart.GetPaymentReference()),
	}, nil
}

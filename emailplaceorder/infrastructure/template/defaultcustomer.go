package template

import (
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure"
	cartDomain "flamingo.me/flamingo-commerce/v3/cart/domain/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	"fmt"
)

// CustomerMail returns mail for customer
func (d *Default) CustomerMail(cart *cartDomain.Cart, payment *placeorder.Payment) (*infrastructure.Mail, error) {

	contentTemplate := `
{{ if .Cart.Purchaser }}
	{{ if .Cart.Purchaser.Address }}
		<h1>Hello {{ .Cart.Purchaser.Address.Salutation }} {{ .Cart.Purchaser.Address.Title }} {{ .Cart.Purchaser.Address.Firstname }} {{ .Cart.Purchaser.Address.MiddleName }} {{ .Cart.Purchaser.Address.Lastname }}</h1>
	{{ end }}
{{ else }}
 	{{ if .Cart.BillingAddress }}
		<h1>Hello {{ .Cart.BillingAddress.Salutation }} {{ .Cart.BillingAddress.Title }} {{ .Cart.BillingAddress.Firstname }} {{ .Cart.BillingAddress.MiddleName }} {{ .Cart.BillingAddress.Lastname }}</h1>
	{{ end }}
{{ end }}
<p>Thank you for your order # {{  .Cart.GetPaymentReference }} </p>
`
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
		Plain:   fmt.Sprintf("Order confirmation for %v / Total: %v", cart.GetContactMail(), d.priceFormat.FormatPrice(cart.GrandTotal())),
		Subject: fmt.Sprintf("Order confirmation - %v", cart.GetPaymentReference()),
	}, nil
}

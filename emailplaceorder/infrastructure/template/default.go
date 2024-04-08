package template

import (
	"bytes"
	"html/template"
	"strings"

	cartDomain "flamingo.me/flamingo-commerce/v3/cart/domain/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	"flamingo.me/flamingo-commerce/v3/price/domain"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/Masterminds/sprig"
	"github.com/vanng822/go-premailer/premailer"

	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure"
)

type (
	//Default email template implementation
	Default struct {
		priceFormat        PriceFormat
		disableCSSInlining bool
		logger             flamingo.Logger
	}

	//MailData that is passed to the template
	MailData struct {
		Cart    *cartDomain.Cart
		Payment *placeorder.Payment
		Link    string
		Logo    string
		Name    string
	}

	//PriceFormat interface
	PriceFormat interface {
		FormatPrice(price domain.Price) string
	}
)

var _ infrastructure.MailTemplate = &Default{}

// Inject dependencies
func (d *Default) Inject(logger flamingo.Logger, priceFormat PriceFormat) {
	d.priceFormat = priceFormat
	d.logger = logger.WithField(flamingo.LogKeyModule, "emailplaceorder")
}

// WrapHTMLTemplate wraps with envelop
func (d *Default) WrapHTMLTemplate(content string, footer string) string {
	template := `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <style type="text/css" rel="stylesheet" media="all">
    /* Base ------------------------------ */
    *:not(br):not(tr):not(html) {
      font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif;
      -webkit-box-sizing: border-box;
      box-sizing: border-box;
    }
    body {
      width: 100% !important;
      height: 100%;
      margin: 0;
      line-height: 1.4;
      background-color: #F2F4F6;
      color: #74787E;
      -webkit-text-size-adjust: none;
    }
    a {
      color: #3869D4;
    }
    /* Layout ------------------------------ */
    .email-wrapper {
      width: 100%;
      margin: 0;
      padding: 0;
      background-color: #F2F4F6;
    }
    .email-content {
      width: 100%;
      margin: 0;
      padding: 0;
    }
    /* Masthead ----------------------- */
    .email-masthead {
      padding: 25px 0;
      text-align: center;
    }
    .email-masthead_logo {
      max-width: 400px;
      border: 0;
    }
    .email-masthead_name {
      font-size: 16px;
      font-weight: bold;
      color: #2F3133;
      text-decoration: none;
      text-shadow: 0 1px 0 white;
    }
    .email-logo {
      max-height: 50px;
    }
    /* Body ------------------------------ */
    .email-body {
      width: 100%;
      margin: 0;
      padding: 0;
      border-top: 1px solid #EDEFF2;
      border-bottom: 1px solid #EDEFF2;
      background-color: #FFF;
    }
    .email-body_inner {
      width: 570px;
      margin: 0 auto;
      padding: 0;
    }
    .email-footer {
      width: 570px;
      margin: 0 auto;
      padding: 0;
      text-align: center;
    }
    .email-footer p {
      color: #AEAEAE;
    }
    .body-action {
      width: 100%;
      margin: 30px auto;
      padding: 0;
      text-align: center;
    }
    .body-dictionary {
      width: 100%;
      overflow: hidden;
      margin: 20px auto 10px;
      padding: 0;
    }
    .body-dictionary dd {
      margin: 0 0 10px 0;
    }
    .body-dictionary dt {
      clear: both;
      color: #000;
      font-weight: bold;
    }
    .body-dictionary dd {
      margin-left: 0;
      margin-bottom: 10px;
    }
    .body-sub {
      margin-top: 25px;
      padding-top: 25px;
      border-top: 1px solid #EDEFF2;
      table-layout: fixed;
    }
    .body-sub a {
      word-break: break-all;
    }
    .content-cell {
      padding: 35px;
    }
    .align-right {
      text-align: right;
    }
    /* Type ------------------------------ */
    h1 {
      margin-top: 0;
      color: #2F3133;
      font-size: 19px;
      font-weight: bold;
    }
    h2 {
      margin-top: 0;
      color: #2F3133;
      font-size: 16px;
      font-weight: bold;
    }
    h3 {
      margin-top: 0;
      color: #2F3133;
      font-size: 14px;
      font-weight: bold;
    }
    blockquote {
      margin: 25px 0;
      padding-left: 10px;
      border-left: 10px solid #F0F2F4;
    }
    blockquote p {
        font-size: 1.1rem;
        color: #999;
    }
    blockquote cite {
        display: block;
        text-align: right;
        color: #666;
        font-size: 1.2rem;
    }
    cite {
      display: block;
      font-size: 0.925rem; 
    }
    cite:before {
      content: "\2014 \0020";
    }
    p {
      margin-top: 0;
      color: #74787E;
      font-size: 16px;
      line-height: 1.5em;
    }
    p.sub {
      font-size: 12px;
    }
    p.center {
      text-align: center;
    }
    table {
      width: 100%;
    }
    th {
      padding: 0px 5px;
      padding-bottom: 8px;
      border-bottom: 1px solid #EDEFF2;
    }
    th p {
      margin: 0;
      color: #9BA2AB;
      font-size: 12px;
    }
    td {
      padding: 10px 5px;
      color: #74787E;
      font-size: 15px;
      line-height: 18px;
    }
    .content {
      align: center;
      padding: 0;
    }
    /* Data table ------------------------------ */
    .data-wrapper {
      width: 100%;
      margin: 0;
      padding: 35px 0;
    }
    .data-table {
      width: 100%;
      margin: 0;
    }
    .data-table th {
      text-align: left;
      padding: 0px 5px;
      padding-bottom: 8px;
      border-bottom: 1px solid #EDEFF2;
    }
    .data-table th p {
      margin: 0;
      color: #9BA2AB;
      font-size: 12px;
    }
    .data-table td {
      padding: 10px 5px;
      color: #74787E;
      font-size: 15px;
      line-height: 18px;
    }
    /* Invite Code ------------------------------ */
    .invite-code {
      display: inline-block;
      padding-top: 20px;
      padding-right: 36px;
      padding-bottom: 16px;
      padding-left: 36px;
      border-radius: 3px;
      font-family: Consolas, monaco, monospace;
      font-size: 28px;
      text-align: center;
      letter-spacing: 8px;
      color: #555;
      background-color: #eee;
    }
    /* Buttons ------------------------------ */
    .button {
      display: inline-block;
      background-color: #3869D4;
      border-radius: 3px;
      color: #ffffff !important;
      font-size: 15px;
      line-height: 45px;
      text-align: center;
      text-decoration: none;
      -webkit-text-size-adjust: none;
      mso-hide: all;
    }
    /*Media Queries ------------------------------ */
    @media only screen and (max-width: 600px) {
      .email-body_inner,
      .email-footer {
        width: 100% !important;
      }
    }
    @media only screen and (max-width: 500px) {
      .button {
        width: 100% !important;
      }
    }
  </style>
</head>
<body>
  <table class="email-wrapper" width="100%" cellpadding="0" cellspacing="0">
    <tr>
      <td class="content">
        <table class="email-content" width="100%" cellpadding="0" cellspacing="0">
          <!-- Logo -->
          <tr>
            <td class="email-masthead">
              <a class="email-masthead_name" href="{{.Link}}" target="_blank">
                {{ if .Logo }}
                  <img src="{{.Logo | url }}" class="email-logo" />
                {{ else }}
                  {{ .Name }}
                {{ end }}
                </a>
            </td>
          </tr>
          <!-- Email Body -->
          <tr>
            <td class="email-body" width="100%">
              <table class="email-body_inner" align="center" width="570" cellpadding="0" cellspacing="0">
                <!-- Body content -->
                <tr>
                  <td class="content-cell">
					###CONTENT###
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td>
              <table class="email-footer" align="center" width="570" cellpadding="0" cellspacing="0">
                <tr>
                  <td class="content-cell">
                    <p class="sub center">
                      ###FOOTER###
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`

	template = strings.Replace(template, "###CONTENT###", content, 1)
	template = strings.Replace(template, "###FOOTER###", footer, 1)
	return template
}

// BasketTemplate helper
func (d *Default) BasketTemplate() string {

	return `
                          <table class="data-wrapper" width="100%" cellpadding="0" cellspacing="0">
                            <tr>
                              <td colspan="2">
                                <table class="data-table" width="100%" cellpadding="0" cellspacing="0">
                                  {{ range $delivery := .Cart.Deliveries }}
									  <tr>
										<th colspan="4">{{ $delivery.DeliveryInfo.Code }} {{ $delivery.DeliveryInfo.Method }}</th>
									  </tr>
									  <tr>
										<td colspan="4">
											{{ if $delivery.DeliveryInfo.DeliveryLocation.UseBillingAddress }}
												Use billing address
											{{ else }}
												{{ if $delivery.DeliveryInfo.DeliveryLocation.Address }}
													{{ with  $delivery.DeliveryInfo.DeliveryLocation.Address }}
` + d.AddressTemplate() + `
													{{ end }}
												{{ end }}
											{{ end }}
										</td>
									  </tr>
									  <tr>
										<th>Item</th>
										<th>Qty</th>
										<th>Price</th>
										<th>Sum</th>
									  </tr>
                                  	{{ range $item := $delivery.Cartitems }}
										<tr>
											<td>{{ $item.ProductName }}</td>
											<td>{{ $item.Qty }}</td>
											<td>{{ priceFormat $item.SinglePriceGross }}</td>
											<td style="text-align:right">{{ priceFormat $item.RowPriceGross }}</td>
										</tr>
                                  	{{ end }}
									{{ if not $delivery.ShippingItem.TotalWithDiscountInclTax.IsZero }}
										<tr>
											<td colspan="3">{{ $delivery.ShippingItem.Title }}</td>
											<td style="text-align:right">{{ priceFormat $delivery.ShippingItem.TotalWithDiscountInclTax }}</td>
										</tr>
									{{ end }}									
                                  {{ end }}
									<tr>
										<th colspan="4"><br>Summary</th>
									</tr>
									{{ if not .Cart.TotalDiscountAmount.IsZero }}
										<tr>
											<td colspan="3">Discounts</td>
											<td style="text-align:right">{{ priceFormat .Cart.TotalDiscountAmount }}</td>
										</tr>
									{{ end }}
									<tr>
										<td colspan="3">Total Tax</td>
										<td style="text-align:right">{{ priceFormat .Cart.SumTotalTaxAmount }}</td>
									</tr>
									<tr>
										<td colspan="3">Grand Total</td>
										<td style="text-align:right">{{ priceFormat .Cart.GrandTotal }}</td>
									</tr>
                                </table>
                              </td>
                            </tr>
                          </table>
`
}

// AddressTemplate helper
func (d *Default) AddressTemplate() string {
	return `
{{ .Salutation }} {{ .Title }} {{ .Firstname }} {{ .MiddleName }} {{ .Lastname }} <br>
{{ if .Company }} {{ .Company }} <br> {{ end }}
{{ .Street }} {{ .StreetNr }} <br>
{{ range $line := .AdditionalAddressLines}}
	{{ $line }} <br>
{{ end }}

{{ .PostCode }} {{ .City }} <br>
{{ if .Country }}{{ .Country }} <br>{{ end }}
{{ if .Telephone }}{{ .Telephone }} <br>{{ end }}
{{ if .Email }}{{ .Email }} <br>{{ end }}
`
}

// GenerateTemplate triggers go template rendering and passes the variables and some additional templatefuncs to it
func (d *Default) GenerateTemplate(cart *cartDomain.Cart, payment *placeorder.Payment, tplt string) (string, error) {

	// Generate the email from Golang template
	// Allow usage of simple function from sprig : https://github.com/Masterminds/sprig

	var templateFuncs = template.FuncMap{
		"url": func(s string) template.URL {
			return template.URL(s)
		},
		"priceFormat": d.priceFormat.FormatPrice,
	}
	t, err := template.New("mail").Funcs(sprig.FuncMap()).Funcs(templateFuncs).Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) }, // Used for keeping comments in generated template
	}).Parse(tplt)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, MailData{Cart: cart, Payment: payment, Logo: "https://raw.githubusercontent.com/i-love-flamingo/flamingo/master/docs/assets/flamingo-logo-only-pink-on-white.png"})
	if err != nil {
		return "", err
	}

	res := b.String()
	if d.disableCSSInlining {
		return res, nil
	}

	// Inlining CSS
	prem, err := premailer.NewPremailerFromString(res, premailer.NewOptions())
	if err != nil {
		return "", err
	}
	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

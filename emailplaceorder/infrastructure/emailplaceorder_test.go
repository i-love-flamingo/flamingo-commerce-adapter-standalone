package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"flamingo.me/flamingo-commerce/v3/cart/domain/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	"flamingo.me/flamingo-commerce/v3/price/domain"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"

	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure"
	"flamingo.me/flamingo-commerce-adapter-standalone/emailplaceorder/infrastructure/template"
)

type mailSenderStub struct {
	called   int
	lastText string
	lastHTML string
}

var _ infrastructure.MailSender = &mailSenderStub{}

func (m *mailSenderStub) Send(credentials infrastructure.Credentials, to string, fromMail string, fromName string, mail *infrastructure.Mail) error {
	m.called++
	m.lastHTML = mail.HTML
	m.lastText = mail.Plain
	return nil
}

type priceFormatter struct{}

func (p *priceFormatter) FormatPrice(price domain.Price) string {
	return fmt.Sprintf("%v %v", price.FloatAmount(), price.Currency())
}

func TestPlaceOrderServiceAdapter_PlaceGuestCart(t *testing.T) {

	t.Run("standard test cart", func(t *testing.T) {
		exampleCart := exampleCart()
		payment := placeorder.Payment{
			Gateway: "test",
			Transactions: []placeorder.Transaction{
				placeorder.Transaction{
					Method:            "testmethod",
					Status:            placeorder.PaymentStatusOpen,
					ValuedAmountPayed: exampleCart.GrandTotal,
					AmountPayed:       exampleCart.GrandTotal,
					TransactionID:     "t1",
				},
			},
			RawTransactionData: nil,
			PaymentID:          "p1",
		}

		adapter, mailsender := initPlaceOrderAdapter()
		po, err := adapter.PlaceGuestCart(context.Background(), &exampleCart, &payment)
		assert.NoError(t, err)
		assert.Len(t, po, 2)
		assert.Contains(t, mailsender.lastHTML, "adrianna@mail.de")
		assert.Contains(t, mailsender.lastHTML, "ProductName 2")
		assert.Contains(t, mailsender.lastHTML, "Opa", "Deliveryaddress not part of mail template")
	})

	t.Run("cart with delivery same as billing ", func(t *testing.T) {
		exampleCart := exampleCart()
		exampleCart.Deliveries[0].DeliveryInfo.DeliveryLocation.UseBillingAddress = true
		payment := placeorder.Payment{
			Gateway: "test",
			Transactions: []placeorder.Transaction{
				placeorder.Transaction{
					Method:            "testmethod",
					Status:            placeorder.PaymentStatusOpen,
					ValuedAmountPayed: exampleCart.GrandTotal,
					AmountPayed:       exampleCart.GrandTotal,
					TransactionID:     "t1",
				},
			},
			RawTransactionData: nil,
			PaymentID:          "p1",
		}

		adapter, mailsender := initPlaceOrderAdapter()
		po, err := adapter.PlaceGuestCart(context.Background(), &exampleCart, &payment)
		assert.NoError(t, err)
		assert.Len(t, po, 2)
		assert.Contains(t, mailsender.lastHTML, "adrianna@mail.de")
		assert.Contains(t, mailsender.lastHTML, "ProductName 2")
		assert.Contains(t, mailsender.lastHTML, "Use billing address", "Same as billing missing")
	})

}

func initPlaceOrderAdapter() (infrastructure.PlaceOrderServiceAdapter, *mailSenderStub) {
	defaultTemplate := &template.Default{}
	defaultTemplate.Inject(flamingo.NullLogger{}, &priceFormatter{})
	mailsender := &mailSenderStub{}
	adapter := infrastructure.PlaceOrderServiceAdapter{}
	adapter.Inject(
		flamingo.NullLogger{},
		defaultTemplate,
		mailsender,
		&struct {
			EmailAddress    string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.emailAddress"`
			FromMail        string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.fromMail,optional"`
			FromName        string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.fromName,optional"`
			SMTPCredentials config.Map `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.credentials"`
		}{
			EmailAddress:    "",
			FromMail:        "",
			FromName:        "",
			SMTPCredentials: nil,
		})
	return adapter, mailsender
}

func exampleCart() cart.Cart {
	return cart.Cart{
		ID:       "",
		EntityID: "",
		BillingAddress: &cart.Address{
			Vat:                    "",
			Firstname:              "Adrianna",
			Lastname:               "Mustermann",
			MiddleName:             "",
			Title:                  "",
			Salutation:             "Mr",
			Street:                 "Musterstraße",
			StreetNr:               "7",
			AdditionalAddressLines: nil,
			Company:                "AOE",
			City:                   "Wiesbaden",
			PostCode:               "65200",
			State:                  "",
			RegionCode:             "",
			Country:                "Germany",
			CountryCode:            "",
			Telephone:              "",
			Email:                  "adrianna@mail.de",
		},
		Purchaser: &cart.Person{
			Address: &cart.Address{
				Vat:                    "",
				Firstname:              "Max",
				Lastname:               "Mustermann",
				MiddleName:             "",
				Title:                  "",
				Salutation:             "Mr",
				Street:                 "Musterstraße",
				StreetNr:               "7",
				AdditionalAddressLines: nil,
				Company:                "AOE",
				City:                   "Wiesbaden",
				PostCode:               "65200",
				State:                  "",
				RegionCode:             "",
				Country:                "Germany",
				CountryCode:            "",
				Telephone:              "",
				Email:                  "mail@mail.de",
			},
			PersonalDetails:      cart.PersonalDetails{},
			ExistingCustomerData: nil,
		},
		Deliveries: []cart.Delivery{
			{
				DeliveryInfo: cart.DeliveryInfo{
					Code:     "delivery",
					Workflow: "",
					Method:   "",
					Carrier:  "",
					DeliveryLocation: cart.DeliveryLocation{
						Type: "",
						Address: &cart.Address{
							Vat:                    "",
							Firstname:              "Opa",
							Lastname:               "Mustermann",
							MiddleName:             "",
							Title:                  "",
							Salutation:             "Mr",
							Street:                 "Musterstraße",
							StreetNr:               "7",
							AdditionalAddressLines: nil,
							Company:                "AOE",
							City:                   "Wiesbaden",
							PostCode:               "65200",
							State:                  "",
							RegionCode:             "",
							Country:                "Germany",
							CountryCode:            "",
							Telephone:              "",
							Email:                  "mail@mail.de",
						},
						UseBillingAddress: false,
						Code:              "",
					},
					AdditionalData:          nil,
					AdditionalDeliveryInfos: nil,
				},
				Cartitems: []cart.Item{
					{
						ID:                     "1",
						ExternalReference:      "",
						MarketplaceCode:        "",
						VariantMarketPlaceCode: "",
						ProductName:            "ProductName",
						SourceID:               "",
						Qty:                    1,
						AdditionalData:         nil,
						SinglePriceGross:       domain.NewFromInt(1190, 100, "€"),
						SinglePriceNet:         domain.NewFromInt(1000, 100, "€"),
						RowPriceGross:          domain.NewFromInt(2380, 100, "€"),
						RowPriceNet:            domain.NewFromInt(2000, 100, "€"),
						RowTaxes:               nil,
						AppliedDiscounts:       nil,
					},
				},
				ShippingItem: cart.ShippingItem{
					Title:            "Express",
					PriceNet:         domain.NewFromInt(1000, 100, "€"),
					TaxAmount:        domain.NewFromInt(190, 100, "€"),
					AppliedDiscounts: nil,
				},
			},
			{
				DeliveryInfo: cart.DeliveryInfo{
					Code:     "pickup",
					Workflow: "pickup",
					Method:   "",
					Carrier:  "",
					DeliveryLocation: cart.DeliveryLocation{
						Type:              "pickup",
						UseBillingAddress: false,
						Code:              "location1",
					},
					AdditionalData:          nil,
					AdditionalDeliveryInfos: nil,
				},
				Cartitems: []cart.Item{
					{
						ID:                     "2",
						ExternalReference:      "",
						MarketplaceCode:        "",
						VariantMarketPlaceCode: "",
						ProductName:            "ProductName 2",
						SourceID:               "",
						Qty:                    1,
						AdditionalData:         nil,
						SinglePriceGross:       domain.NewFromInt(1190, 100, "€"),
						SinglePriceNet:         domain.NewFromInt(1000, 100, "€"),
						RowPriceGross:          domain.NewFromInt(2380, 100, "€"),
						RowPriceNet:            domain.NewFromInt(2000, 100, "€"),
						RowTaxes:               nil,
						AppliedDiscounts:       nil,
					},
					{
						ID:                     "3",
						ExternalReference:      "",
						MarketplaceCode:        "",
						VariantMarketPlaceCode: "",
						ProductName:            "ProductName 3",
						SourceID:               "",
						Qty:                    1,
						AdditionalData:         nil,
						SinglePriceGross:       domain.NewFromInt(1190, 100, "€"),
						SinglePriceNet:         domain.NewFromInt(1000, 100, "€"),
						RowPriceGross:          domain.NewFromInt(2380, 100, "€"),
						RowPriceNet:            domain.NewFromInt(2000, 100, "€"),
						RowTaxes:               nil,
						AppliedDiscounts:       nil,
					},
				},
			},
		},
		AdditionalData:             cart.AdditionalData{},
		PaymentSelection:           nil,
		BelongsToAuthenticatedUser: false,
		AuthenticatedUserID:        "",
		AppliedCouponCodes:         nil,
		DefaultCurrency:            "",
		Totalitems:                 nil,
		AppliedGiftCards:           nil,
	}

}

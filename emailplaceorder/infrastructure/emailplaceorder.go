package infrastructure

import (
	"context"
	"errors"
	"flamingo.me/flamingo/v3/framework/config"
	"fmt"
	"net/smtp"
	"net/textproto"
	"strings"

	cartDomain "flamingo.me/flamingo-commerce/v3/cart/domain/cart"
	"flamingo.me/flamingo-commerce/v3/cart/domain/placeorder"
	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/jordan-wright/email"
)

type (
	// PlaceOrderServiceAdapter provides an implementation of the Service as email adapter
	PlaceOrderServiceAdapter struct {
		emailAddress string
		fromMail     string
		fromName     string
		websiteURL   string
		logger       flamingo.Logger
		credentials  Credentials
		mailSender   MailSender
		mailTemplate MailTemplate
	}

	// Credentials defines the smtp credentials provided for the SendService
	Credentials struct {
		Password string
		Server   string
		Port     string
		User     string
	}

	//MailSender interface
	MailSender interface {
		Send(credentials Credentials, to string, fromMail string, fromName string, mail *Mail) error
	}

	//MailTemplate interface
	MailTemplate interface {
		AdminMail(cart *cartDomain.Cart, payment *placeorder.Payment) (*Mail, error)
		CustomerMail(cart *cartDomain.Cart, payment *placeorder.Payment) (*Mail, error)
	}

	//Mail representation
	Mail struct {
		HTML    string
		Plain   string
		Subject string
	}

	//DefaultMailSender implementation of MailSender
	DefaultMailSender struct {
		logger flamingo.Logger
	}
)

var (
	_ placeorder.Service = new(PlaceOrderServiceAdapter)
	_ MailSender         = new(DefaultMailSender)
)

// Inject dependencies
func (e *PlaceOrderServiceAdapter) Inject(logger flamingo.Logger, mailTemplate MailTemplate, mailSender MailSender,
	config *struct {
		EmailAddress    string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.emailAddress"`
		FromMail        string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.fromMail,optional"`
		FromName        string     `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.fromName,optional"`
		SMTPCredentials config.Map `inject:"config:flamingoCommerceAdapterStandalone.emailplaceorder.credentials"`
	}) {
	e.mailTemplate = mailTemplate
	e.mailSender = mailSender
	e.logger = logger.WithField("module", "flamingo-commerce-adapter-standalone.emailplaceorder").WithField("category", "emailplaceorder")
	if config != nil {
		err := config.SMTPCredentials.MapInto(&e.credentials)
		if err != nil {
			e.logger.Error(err)
		}
		e.emailAddress = config.EmailAddress
		e.fromMail = config.FromMail
		e.fromName = config.FromName
	}
}

// PlaceGuestCart places a guest cart as order email
func (e *PlaceOrderServiceAdapter) PlaceGuestCart(ctx context.Context, cart *cartDomain.Cart, payment *placeorder.Payment) (placeorder.PlacedOrderInfos, error) {
	return e.placeOrder(ctx, cart, payment)
}

// PlaceCustomerCart places a customer cart as order email
func (e *PlaceOrderServiceAdapter) PlaceCustomerCart(ctx context.Context, auth auth.Identity, cart *cartDomain.Cart, payment *placeorder.Payment) (placeorder.PlacedOrderInfos, error) {
	return e.placeOrder(ctx, cart, payment)
}

// placeOrder
func (e *PlaceOrderServiceAdapter) placeOrder(ctx context.Context, cart *cartDomain.Cart, payment *placeorder.Payment) (placeorder.PlacedOrderInfos, error) {
	err := e.checkPayment(cart, payment)
	if err != nil {
		return nil, err
	}
	var placedOrders placeorder.PlacedOrderInfos
	for _, del := range cart.Deliveries {
		placedOrders = append(placedOrders, placeorder.PlacedOrderInfo{
			OrderNumber:  cart.ID,
			DeliveryCode: del.DeliveryInfo.Code,
		})
	}
	err = e.sendAdminMail(cart, payment, placedOrders)
	if err != nil {
		return nil, err
	}

	err = e.sendCustomerMail(cart, payment, placedOrders)
	if err != nil {
		return nil, err
	}
	return placedOrders, nil
}

// checkPayment
func (e *PlaceOrderServiceAdapter) checkPayment(cart *cartDomain.Cart, payment *placeorder.Payment) error {
	if payment == nil && cart.GrandTotal().IsPositive() {
		return errors.New("No valid Payment given")
	}
	if cart.GrandTotal().IsPositive() {
		totalPrice, err := payment.TotalValue()
		if err != nil {
			return err
		}
		if !totalPrice.Equal(cart.GrandTotal()) {
			return errors.New("Payment Total does not match with Grandtotal")
		}
	}
	return nil
}

// ReserveOrderID returns the reserved order id
func (e *PlaceOrderServiceAdapter) ReserveOrderID(ctx context.Context, cart *cartDomain.Cart) (string, error) {
	return cart.ID, nil
}

// CancelGuestOrder cancels a guest order
func (e *PlaceOrderServiceAdapter) CancelGuestOrder(ctx context.Context, orderInfos placeorder.PlacedOrderInfos) error {
	orderNumbers := []string{}
	for _, poi := range orderInfos {
		orderNumbers = append(orderNumbers, poi.OrderNumber)
	}
	return e.sendMail(e.emailAddress, &Mail{Subject: fmt.Sprintf("Guest Order(s) %v  canceled", strings.Join(orderNumbers, ";"))})
}

// CancelCustomerOrder cancels a customer order
func (e *PlaceOrderServiceAdapter) CancelCustomerOrder(ctx context.Context, orderInfos placeorder.PlacedOrderInfos, auth auth.Identity) error {
	orderNumbers := []string{}
	for _, poi := range orderInfos {
		orderNumbers = append(orderNumbers, poi.OrderNumber)
	}
	return e.sendMail(e.emailAddress, &Mail{Subject: fmt.Sprintf("Customer Order(s) %v  canceled", strings.Join(orderNumbers, ";"))})

}

func (e *PlaceOrderServiceAdapter) sendAdminMail(cart *cartDomain.Cart, payment *placeorder.Payment, placedOrders placeorder.PlacedOrderInfos) error {
	mail, err := e.mailTemplate.AdminMail(cart, payment)
	if err != nil {
		return err
	}
	return e.sendMail(e.emailAddress, mail)
}

func (e *PlaceOrderServiceAdapter) sendCustomerMail(cart *cartDomain.Cart, payment *placeorder.Payment, placedOrders placeorder.PlacedOrderInfos) error {
	mail, err := e.mailTemplate.CustomerMail(cart, payment)
	if err != nil {
		return err
	}
	return e.sendMail(cart.GetContactMail(), mail)
}

func (e *PlaceOrderServiceAdapter) sendMail(to string, mail *Mail) error {
	err := e.mailSender.Send(e.credentials, to, e.fromMail, e.fromName, mail)
	if err != nil {
		e.logger.Error(err)
		return err
	}
	return nil
}

//Inject dep
func (m *DefaultMailSender) Inject(logger flamingo.Logger) {
	m.logger = logger.WithField(flamingo.LogKeyModule, "flamingo-commerce-adapter-standalone.emailplaceorder").WithField("category", "emailplaceorder")
}

//Send mail
func (m *DefaultMailSender) Send(credentials Credentials, to string, fromMail string, fromName string, mail *Mail) error {
	err := credentials.Validate()
	if err != nil {
		return err
	}
	email := &email.Email{
		To:      []string{to},
		From:    fmt.Sprintf("%v <%v>", fromName, fromMail),
		Subject: mail.Subject,
		Text:    []byte(mail.Plain),
		HTML:    []byte(mail.HTML),
		Headers: textproto.MIMEHeader{},
	}
	m.logger.Debugf("Try to send mail. Via SMTP Server: %v / TO: %v ; From : %v %v", credentials.Server, to, fromName, fromMail)
	return email.Send(fmt.Sprintf("%v:%v", credentials.Server, credentials.Port), smtp.PlainAuth("", credentials.User, credentials.Password, credentials.Server))
}

//Validate helper
func (s *Credentials) Validate() error {
	if s.Server == "" {
		return errors.New("Credentials missing server")
	}
	if s.Password == "" {
		return errors.New("Credentials missing password")
	}
	if s.User == "" {
		return errors.New("Credentials missing user")
	}
	if s.Port == "" {
		return errors.New("Credentials missing port")
	}
	return nil
}

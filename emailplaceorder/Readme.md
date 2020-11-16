# E-Mail place order adapter

Adapter that sends basic emails as soon as a customer places an order. Both the admin/owner of the shop as well as the
customer get an email confirmation.

## `MailSender` Port

To deliver an email a custom `MailSender` implementation can be used. The module comes with a default implementation
that relies on basic SMTP functionallity.

## `MailTemplate` Port

Besides, the actual mail delivery there is also a port for defining the used email templates. There is a default
implementation as well, see `infrastructure/template` for details.

## Configuration

There are various configurations that can be set. In addition, the default SMTP delivery needs some credentials to work.

```yaml
flamingoCommerceAdapterStandalone:
  emailplaceorder:
    emailAddress: fullfilment@flamingo-shop.example
    fromMail: no-reply@flamingo-shop.example
    fromName: Flamingo Shop
    credentials:
      server: smtp.example.com
      port: 587
      user: user
      password: %%ENV:SMTP_PASSWORD%%
```


 

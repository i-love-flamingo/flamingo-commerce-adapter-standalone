module flamingo.me/flamingo-commerce-adapter-standalone

require (
	flamingo.me/dingo v0.1.5
	flamingo.me/flamingo-commerce/v3 v3.0.0-beta.1
	flamingo.me/flamingo/v3 v3.0.0-beta.2.0.20190515120627-9cabe248cf01
	github.com/disintegration/imaging v1.5.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 // indirect
)

replace flamingo.me/flamingo/v3 => ../flamingo

replace flamingo.me/flamingo-commerce/v3 => ../flamingo-commerce

replace flamingo.me/form => ../form

replace flamingo.me/flamingo-commerce-adapter-standalone => ../flamingo-commerce-adapter-standalone

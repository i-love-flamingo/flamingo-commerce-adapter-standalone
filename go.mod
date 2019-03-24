module flamingo.me/flamingo-commerce-adapter-standalone

require (
	flamingo.me/dingo v0.1.4
	flamingo.me/flamingo-commerce/v3 v3.0.0-alpha5
	flamingo.me/flamingo/v3 v3.0.0
	github.com/disintegration/imaging v1.5.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 // indirect
)

replace flamingo.me/flamingo/v3 => ../flamingo

replace flamingo.me/flamingo-commerce/v3 => ../flamingo-commerce

replace flamingo.me/form => ../form

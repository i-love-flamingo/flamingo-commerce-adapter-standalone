CONTEXT?=dev
.PHONY: up localup update test

up:
	rm -rf vendor/
	dep ensure -v -vendor-only

update:
	rm -rf vendor/
	dep ensure -v -update flamingo.me/flamingo
	dep ensure -v -update flamingo.me/flamingo-commerce

localup: up local
	
local:
	rm -rf vendor/flamingo.me/flamingo
	ln -sf ../../../flamingo vendor/flamingo.me/flamingo
	rm -rf vendor/flamingo.me/flamingo/vendor
	rm -rf vendor/flamingo.me/flamingo-commerce
	ln -sf ../../../flamingo vendor/flamingo.me/flamingo-commerce
	rm -rf vendor/flamingo.me/flamingo-commerce/vendor
	
test:
	go test -v ./...
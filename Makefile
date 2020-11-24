.PHONY: local unlocal test
REPLACE?=-replace flamingo.me/flamingo/v3=../flamingo -replace flamingo.me/flamingo-commerce/v3=../flamingo-commerce
DROPREPLACE?=-dropreplace flamingo.me/flamingo/v3 -dropreplace flamingo.me/flamingo-commerce/v3

local:
	git config filter.gomod-commerceadapter-standalone.smudge 'go mod edit -fmt -print $(REPLACE) /dev/stdin'
	git config filter.gomod-commerceadapter-standalone.clean 'go mod edit -fmt -print $(DROPREPLACE) /dev/stdin'
	git config filter.gomod-commerceadapter-standalone.required true
	go mod edit -fmt $(REPLACE)

unlocal:
	git config filter.gomod-commerceadapter-standalone.smudge ''
	git config filter.gomod-commerceadapter-standalone.clean ''
	git config filter.gomod-commerceadapter-standalone.required false
	go mod edit -fmt $(DROPREPLACE)

test:
	go test -race -v ./...
	gofmt -l -e -d .
	golint ./...
	find . -type f -name '*.go' | xargs go run github.com/client9/misspell/cmd/misspell -error
	find . -type f -name '*.md' | xargs go run github.com/client9/misspell/cmd/misspell -error
	ineffassign .

integrationtest:
	go test -race -v ./test/integrationtest/... -tags=integration

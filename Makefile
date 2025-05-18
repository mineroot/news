.PHONY: test vendor update-deps

default: test

test:
	go test -race ./...

mockery:
	go tool mockery

vendor:
	go mod tidy
	go mod vendor

update-deps:
	go get -u ./...
	$(MAKE) vendor
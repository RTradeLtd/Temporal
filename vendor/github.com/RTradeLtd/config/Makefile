all: build

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: lint
lint:
	golint $(go list ./... | grep -v /vendor/)

.PHONY: config
config:
	go run cmd/cfggen/main.go

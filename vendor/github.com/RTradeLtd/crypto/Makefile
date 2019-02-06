all: deps install

.PHONY: deps
deps:
	dep ensure

.PHONY: install
install:
	go install ./cmd/temporal-crypto

GO=env GO111MODULE=on go
GONOMOD=env GO111MODULE=off go
all: build

.PHONY: deps
deps:
	$(GO) mod vendor
	$(GO) mod tidy
	
.PHONY: build
build: deps
	go build ./...

.PHONY: testenv
testenv:
	( cd testenv ; make testenv )

.PHONY: test
test: vendor
	go test -race -cover ./...

.PHONY: lint
lint: vendor
	golint $(go list ./... | grep -v /vendor/)

.PHONY: clean
clean:
	( cd testenv ; make clean )

.PHONY: check
check:
	go vet ./...
	go test -run xxxx ./...

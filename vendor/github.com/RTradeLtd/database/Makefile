COMPOSECOMMAND=env ADDR_NODE1=1 ADDR_NODE2=2 docker-compose -f testenv/docker-compose.yml

all: build

.PHONY: vendor
vendor:
	dep ensure

.PHONY: build
build: vendor
	go build ./...

.PHONY: testenv
testenv:
	$(COMPOSECOMMAND) up -d postgres

.PHONY: clean
clean:
	$(COMPOSECOMMAND) down

.PHONY: test
test: vendor
	go test -race -cover ./...

.PHONY: lint
lint: vendor
	golint $(go list ./... | grep -v /vendor/)

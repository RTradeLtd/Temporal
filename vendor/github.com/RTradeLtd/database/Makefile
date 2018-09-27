TESTCONFIG=https://raw.githubusercontent.com/RTradeLtd/Temporal/V2/test/config.json
TESTCOMPOSE=https://raw.githubusercontent.com/RTradeLtd/Temporal/V2/test/docker-compose.yml

COMPOSECOMMAND=env ADDR_NODE1=1 ADDR_NODE2=2 docker-compose -f test/docker-compose.yml

all: build

.PHONY: vendor
vendor:
	dep ensure

.PHONY: build
build: vendor
	go build ./...

.PHONY: testenv
testenv:
	mkdir -p test
	curl $(TESTCONFIG) --output test/config.json
	curl $(TESTCOMPOSE) --output test/docker-compose.yml
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

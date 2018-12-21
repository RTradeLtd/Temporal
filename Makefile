TEMPORALVERSION=`git describe --tags`

all: check cli

# Build temporal if binary is not already present
temporal:
	@make cli

# List all commands
.PHONY: ls
ls:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

# Installs Temporal to GOBIN
.PHONY: install
install: cli
	@echo "=================== installing Temporal CLI ==================="
	go install -ldflags "-X main.Version=$(TEMPORALVERSION)" cmd/temporal
	@echo "===================          done           ==================="

# Run simple checks
.PHONY: check
check:
	@echo "===================      running checks     ==================="
	git submodule update --init
	go vet ./...
	@echo "Executing dry run of tests..."
	@go test -run xxxx ./...
	@echo "===================          done           ==================="

# Build Temporal
.PHONY: cli
cli:
	@echo "===================  building Temporal CLI  ==================="
	rm -f temporal
	go build -ldflags "-X main.Version=$(TEMPORALVERSION)" ./cmd/temporal
	@echo "===================          done           ==================="

# Static analysis and style checks
.PHONY: lint
lint:
	go fmt ./...
	golint $(GOFILES)
	# Shellcheck disabled for now - too much to fix
	# shellcheck **/*.sh(e[' [[ ! `echo "$REPLY" | grep "vendor/" ` ]]'])

# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	( cd testenv ; make testenv )
	@echo "Running migrations..."
	@env CONFIG_DAG=./testenv/config.json go run cmd/temporal/main.go migrate-insecure
	make api-user
	make api-admin
	@echo "===================          done           ==================="

# Shut down testenv
.PHONY: stop-testenv
stop-testenv:
	@echo "===================  shutting down test env ==================="
	( cd testenv ; make stop-testenv )
	@echo "===================          done           ==================="

# Execute short tests
.PHONY: test
test: check
	@echo "===================  executing short tests  ==================="
	go test -race -cover -short ./...
	@echo "===================          done           ==================="

# Execute all tests
.PHONY: test
test-all: check
	@echo "===================   executing all tests   ==================="
	go test -race -cover ./...
	@echo "===================          done           ==================="

# Remove assets
.PHONY: clean
clean: stop-testenv
	@echo "=================== cleaning up temp assets ==================="
	@echo "Removing binary..."
	@rm -f temporal
	( cd testenv ; make clean )
	@echo "===================          done           ==================="

# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	# Nuke vendor directory
	rm -rf vendor

	# Update standard dependencies
	dep ensure -v

	# Generate IPFS dependencies
	rm -rf vendor/github.com/ipfs/go-ipfs
	git clone https://github.com/ipfs/go-ipfs.git vendor/github.com/ipfs/go-ipfs
	( cd vendor/github.com/ipfs/go-ipfs ; git checkout $(IPFSVERSION) ; gx install --local --nofancy )
	mv vendor/github.com/ipfs/go-ipfs/vendor/* vendor
	
	# Remove problematic dependencies
	find . -name test-vectors -type d -exec rm -r {} +
	@echo "===================          done           ==================="

# Build CLI binary release
.PHONY: release-cli
release-cli:
	@echo "===================   cross-compiling CLI   ==================="
	@bash .scripts/cli.sh
	@echo "===================          done           ==================="

# Build docker release
.PHONY: docker
docker:
	@echo "===================  building docker image  ==================="
	@docker build --build-arg TEMPORALVERSION=$(TEMPORALVERSION) \
		-t rtradetech/temporal:$(TEMPORALVERSION) .
	@echo "===================          done           ==================="

# Build docker release and push to repository
.PHONY: release-docker
release-docker: docker
	@echo "===================  building docker image  ==================="
	@docker push rtradetech/temporal:$(TEMPORALVERSION)
	@echo "===================          done           ==================="

# Run development API
.PHONY: api
api:
	CONFIG_DAG=./testenv/config.json go run cmd/temporal/main.go api

USER=testuser
PASSWORD=admin
EMAIL=test@email.com

.PHONY: api-user
api-user:
	CONFIG_DAG=./testenv/config.json go run cmd/temporal/main.go user $(USER) $(PASSWORD) $(EMAIL)

.PHONY: api-admin
api-admin:
	CONFIG_DAG=./testenv/config.json go run cmd/temporal/main.go admin $(USER)

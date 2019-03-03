TEMPORALVERSION=`git describe --tags`
TEMPORALDEVFLAGS=-config ./testenv/config.json -db.no_ssl -dev
IPFSVERSION=v0.4.19

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

.PHONY: postgres
WAIT=3
postgres:
	( cd testenv ; make postgres )

# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	( cd testenv ; make testenv )
	@echo "Running migrations..."
	go run cmd/temporal/main.go -config ./testenv/config.json --db.no_ssl migrate
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

# Rebuild generate code
.PHONY: gen
gen:
	@echo "===================    regenerating code    ==================="
	counterfeiter -o ./mocks/orchestrator.mock.go \
		./vendor/github.com/RTradeLtd/grpc/nexus ServiceClient
	counterfeiter -o ./mocks/lens.mock.go \
		./vendor/github.com/RTradeLtd/grpc/lensv2 LensV2Client
	counterfeiter -o ./mocks/eth.mock.go \
		./vendor/github.com/RTradeLtd/grpc/pay SignerClient
	counterfeiter -o ./mocks/rtfs.mock.go \
		./vendor/github.com/RTradeLtd/rtfs Manager
	@echo "===================          done           ==================="

# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	# Nuke vendor directory
	rm -rf vendor

	# rebuild dependencies
	dep ensure -v $(DEPFLAGS)

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

# download and setup gvisor runtime
.PHONY: gvisor
gvisor:
	wget https://storage.googleapis.com/gvisor/releases/nightly/latest/runsc -O tmp/runsc
	wget https://storage.googleapis.com/gvisor/releases/nightly/latest/runsc.sha512 -O tmp/runsc.sha512
	sha512sum -c tmp/runsc.sha512
	chmod a+x tmp/runsc
	sudo mv tmp/runsc /usr/local/bin
	sudo cp setup/configs/docker/daemon_passthrough.json /etc/docker/daemon.json
	sudo systemctl restart docker

# Run development API
.PHONY: api
api:
	go run cmd/temporal/main.go $(TEMPORALDEVFLAGS) api

USER=testuser
PASSWORD=admin
EMAIL=test@email.com

.PHONY: api-user
api-user:
	go run cmd/temporal/main.go $(TEMPORALDEVFLAGS) user $(USER) $(PASSWORD) $(EMAIL)

.PHONY: api-admin
api-admin:
	go run cmd/temporal/main.go $(TEMPORALDEVFLAGS) admin $(USER)

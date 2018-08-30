all: check Temporal

# List all commands
.PHONY: ls
ls:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

# Run simple checks
.PHONY: check
check:
	@echo "===================      running checks     ==================="
	go vet ./...
	go test -run xxxx ./...
	@echo "===================          done           ==================="

# Build Temporal
Temporal:
	go build ./...

# Static analysis and style checks
.PHONY: lint
lint:
	go fmt ./...
	# Shellcheck disabled for now - too much to fix
	# shellcheck **/*.sh(e[' [[ ! `echo "$REPLY" | grep "vendor/" ` ]]'])

# Set up test environment
.PHONY: testenv
testenv:
	docker-compose -f test/docker-compose.yml up 

# Execute tests
.PHONY: test
test: check
	@echo "===================     executing tests     ==================="
	go test -race -cover ./...
	@echo "===================          done           ==================="

# Remove assets
.PHONY: clean
clean:
	rm -f Temporal
	docker-compose -f test/docker-compose.yml rm -f -s -v

# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	# Nuke vendor directory
	rm -rf vendor

	# Run dep ensure first, as it clears the vendor directory upon completion
	dep ensure -v

	# Install required IPFS GX dependencies
	go get -u github.com/whyrusleeping/gx
	while read in; do gx get -o ./vendor/gx/ipfs/"$$in" "$$in"; done < ipfs_deps.txt

	# Vendor IPFS. The previous step resolves the required GX dependencies
	# rather than using `make install` in go-ipfs, which installs everything -
	# Temporal only requires a subset for the time being.
	go get -u github.com/ipfs/go-ipfs
	mkdir -p ./vendor/github.com/ipfs/go-ipfs
	cp -r $(GOPATH)/src/github.com/ipfs/go-ipfs ./vendor/github.com/ipfs/go-ipfs

	# Vendor ethereum - this step is required for some of the cgo components, as
	# dep doesn't seem to resolve them
	go get -u github.com/ethereum/go-ethereum
	cp -r $(GOPATH)/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1 \
  	./vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/
	@echo "===================          done           ==================="

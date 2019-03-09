PKGS=core auth store ipfs

all: deps

.PHONY: deps
deps:
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get github.com/golang/protobuf/protoc-gen-go
	npm install -g redoc-cli

.PHONY: proto
proto:
	$(foreach var,$(PKGS),$(MAKE) gen-pkg PKG=$(var);)

.PHONY: gen-pkg
gen-pkg:
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		${PKG}/${PKG}.proto
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:. \
		${PKG}/${PKG}.proto

.PHONY: swagger
swagger:
	$(foreach var,$(PKGS),$(MAKE) swagger-pkg PKG=$(var);)

.PHONY: swagger-pkg
swagger-pkg:
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:. \
		${PKG}/${PKG}.proto

.PHONY: docs
docs: swagger
	$(foreach var,$(PKGS),$(MAKE) docs-pkg PKG=$(var);)

.PHONY: docs-pkg
docs-pkg:
	redoc-cli bundle ${PKG}/${PKG}.swagger.json -o docs/v3/${PKG}/index.html

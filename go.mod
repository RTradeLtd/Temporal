module github.com/RTradeLtd/Temporal

go 1.14

require (
	github.com/RTradeLtd/ChainRider-Go v1.0.8
	github.com/RTradeLtd/cmd/v2 v2.1.0
	github.com/RTradeLtd/config/v2 v2.2.0
	github.com/RTradeLtd/crypto/v2 v2.1.1
	github.com/RTradeLtd/database/v2 v2.7.4-rc1
	github.com/RTradeLtd/entropy-mnemonics v0.0.0-20170316012907-7b01a644a636
	github.com/RTradeLtd/go-ipfs-api v0.0.0-20190522213636-8e3700e602fd
	github.com/RTradeLtd/gpaginator v0.0.4
	github.com/RTradeLtd/grpc v0.0.0-20190528193535-5184ecc77228
	github.com/RTradeLtd/kaas/v2 v2.1.3
	github.com/RTradeLtd/rtfs/v2 v2.1.2
	github.com/RTradeLtd/rtns v0.0.19
	github.com/appleboy/gin-jwt v2.3.1+incompatible
	github.com/appleboy/gofight/v2 v2.1.1 // indirect
	github.com/aviddiviner/gin-limit v0.0.0-20170918012823-43b5f79762c1
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23 // indirect
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dvwright/xss-mw v0.0.0-20191029162136-7a0dab86d8f6
	github.com/fatih/color v1.9.0 // indirect
	github.com/gcash/bchutil v0.0.0-20191012211144-98e73ec336ba
	github.com/gcash/bchwallet v0.8.2
	github.com/gin-contrib/secure v0.0.1
	github.com/gin-gonic/gin v1.6.2
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/ipfs/go-bitswap v0.1.11 // indirect
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-datastore v0.3.1
	github.com/ipfs/go-ds-badger v0.2.1-0.20191209122420-222c6d760ad4
	github.com/ipfs/go-ipfs-addr v0.0.1
	github.com/ipfs/go-ipfs-blockstore v0.1.4-0.20200204183011-fa2cbe80b729 // indirect
	github.com/ipfs/go-ipfs-chunker v0.0.4 // indirect
	github.com/ipfs/go-ipfs-provider v0.3.0 // indirect
	github.com/ipfs/go-log v1.0.2 // indirect
	github.com/ipfs/go-merkledag v0.3.1 // indirect
	github.com/ipfs/go-path v0.0.7
	github.com/ipfs/go-unixfs v0.2.4 // indirect
	github.com/ipfs/ipfs-cluster v0.12.1
	github.com/jinzhu/gorm v1.9.11
	github.com/jszwec/csvutil v1.2.1
	github.com/jung-kurt/gofpdf v1.16.2 // indirect
	github.com/libp2p/go-libp2p-connmgr v0.2.1 // indirect
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/libp2p/go-libp2p-pubsub v0.2.5 // indirect
	github.com/libp2p/go-libp2p-quic-transport v0.2.3 // indirect
	github.com/libp2p/go-libp2p-tls v0.1.3 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2 // indirect
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/multiformats/go-multihash v0.0.13
	github.com/pkg/errors v0.9.1 // indirect
	github.com/polydawn/refmt v0.0.0-20190807091052-3d65705ee9f1 // indirect
	github.com/prometheus/client_golang v1.4.1 // indirect
	github.com/rs/cors v1.7.0
	github.com/semihalev/gin-stats v0.0.0-20180505163755-30fdcbbd3533
	github.com/sendgrid/rest v2.4.1+incompatible
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stripe/stripe-go v60.0.1+incompatible
	github.com/tv42/httpunix v0.0.0-20191220191345-2ba4b9c3382c // indirect
	go.bobheadxi.dev/res v0.2.0
	go.bobheadxi.dev/zapx/zapx v0.6.8
	go.bobheadxi.dev/zapx/ztest v0.6.4
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200208060501-ecb85df21340 // indirect
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	golang.org/x/tools v0.0.0-20200207224406-61798d64f025 // indirect
	google.golang.org/api v0.5.0 // indirect
	google.golang.org/grpc v1.25.1
	gopkg.in/appleboy/gofight.v2 v2.0.0-00010101000000-000000000000 // indirect
	gopkg.in/dgrijalva/jwt-go.v3 v3.2.0
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

replace github.com/dgraph-io/badger v2.0.0-rc.2+incompatible => github.com/dgraph-io/badger v1.6.0-rc1

replace gopkg.in/appleboy/gofight.v2 => github.com/appleboy/gofight/v2 v2.1.2-0.20190917094147-9fdcf0fe61e5

replace github.com/golangci/golangci-lint => github.com/golangci/golangci-lint v1.18.0

replace github.com/go-critic/go-critic v0.0.0-20181204210945-ee9bf5809ead => github.com/go-critic/go-critic v0.3.5-0.20190526074819-1df300866540

replace github.com/ipfs/go-merkledag v0.3.1 => github.com/ipfs/go-merkledag v0.0.3

module github.com/RTradeLtd/Temporal

go 1.13

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/RTradeLtd/ChainRider-Go v1.0.8
	github.com/RTradeLtd/cmd/v2 v2.1.0
	github.com/RTradeLtd/config/v2 v2.2.0
	github.com/RTradeLtd/crypto/v2 v2.1.1
	github.com/RTradeLtd/database/v2 v2.6.2
	github.com/RTradeLtd/entropy-mnemonics v0.0.0-20170316012907-7b01a644a636
	github.com/RTradeLtd/go-ipfs-api v0.0.0-20190522213636-8e3700e602fd
	github.com/RTradeLtd/grpc v0.0.0-20190528193535-5184ecc77228
	github.com/RTradeLtd/kaas/v2 v2.1.2
	github.com/RTradeLtd/rtfs/v2 v2.1.2
	github.com/RTradeLtd/rtns v1.0.2-rc1
	github.com/appleboy/gin-jwt v2.3.1+incompatible
	github.com/appleboy/gofight/v2 v2.1.1 // indirect
	github.com/aviddiviner/gin-limit v0.0.0-20170918012823-43b5f79762c1
	github.com/bobheadxi/res v0.1.0
	github.com/bobheadxi/zapx v0.2.0
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23 // indirect
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dvwright/xss-mw v0.0.0-20191029162136-7a0dab86d8f6
	github.com/gcash/bchutil v0.0.0-20191012211144-98e73ec336ba
	github.com/gcash/bchwallet v0.8.2
	github.com/gin-contrib/secure v0.0.0-20190301062601-f9a5befa6106
	github.com/gin-gonic/gin v1.4.1-0.20190911101039-b562fed3aa28
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/google/uuid v1.1.1
	github.com/hsanjuan/go-libp2p-gostream v0.1.0 // indirect
	github.com/ipfs/go-cid v0.0.3
	github.com/ipfs/go-datastore v0.1.1
	github.com/ipfs/go-ds-badger v0.0.7
	github.com/ipfs/go-ipfs-addr v0.0.1
	github.com/ipfs/go-ipfs-config v0.0.6 // indirect
	github.com/ipfs/go-ipns v0.0.1
	github.com/ipfs/go-path v0.0.7
	github.com/ipfs/ipfs-cluster v0.9.0-rc1.0.20190527122736-35860d337f6b
	github.com/jinzhu/gorm v1.9.11
	github.com/libp2p/go-libp2p v0.4.2
	github.com/libp2p/go-libp2p-core v0.2.5
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-interface-pnet v0.1.0 // indirect
	github.com/libp2p/go-libp2p-kad-dht v0.1.1
	github.com/libp2p/go-libp2p-peerstore v0.1.4
	github.com/libp2p/go-libp2p-pubsub v0.1.0 // indirect
	github.com/libp2p/go-libp2p-raft v0.0.4 // indirect
	github.com/libp2p/go-libp2p-record v0.1.0
	github.com/microcosm-cc/bluemonday v1.0.2 // indirect
	github.com/multiformats/go-multiaddr v0.1.2
	github.com/multiformats/go-multihash v0.0.8
	github.com/prometheus/common v0.4.1 // indirect
	github.com/prometheus/procfs v0.0.2 // indirect
	github.com/rs/cors v1.6.0
	github.com/semihalev/gin-stats v0.0.0-20180505163755-30fdcbbd3533
	github.com/sendgrid/rest v2.4.1+incompatible
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stripe/stripe-go v60.0.1+incompatible
	go.uber.org/zap v1.11.0
	google.golang.org/grpc v1.25.1
	gopkg.in/appleboy/gofight.v2 v2.0.0-00010101000000-000000000000 // indirect
	gopkg.in/dgrijalva/jwt-go.v3 v3.2.0
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

replace github.com/dgraph-io/badger v2.0.0-rc.2+incompatible => github.com/dgraph-io/badger v1.6.0-rc1

replace gopkg.in/appleboy/gofight.v2 => github.com/appleboy/gofight/v2 v2.1.2-0.20190917094147-9fdcf0fe61e5

replace github.com/golangci/golangci-lint => github.com/golangci/golangci-lint v1.18.0

replace github.com/go-critic/go-critic v0.0.0-20181204210945-ee9bf5809ead => github.com/go-critic/go-critic v0.3.5-0.20190526074819-1df300866540

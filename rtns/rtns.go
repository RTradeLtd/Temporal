package rtns

/*
IPNS related functinonality for temporal
*/
import (
	"errors"
	"fmt"
	"time"

	ipns "github.com/ipfs/go-ipns"
	pb "github.com/ipfs/go-ipns/pb"
	lci "github.com/libp2p/go-libp2p-crypto"
)

// IpnsManager is used to interface with IPNS
type IpnsManager struct {
	PrivateKey lci.PrivKey
	PublicKey  lci.PubKey
	KeyType    int
}

// InitializeWithNewKey is used to generate our ipns manager
// with a newly generated random key
func InitializeWithNewKey() (*IpnsManager, error) {
	manager := IpnsManager{}
	fmt.Println("generating key")
	err := manager.GenerateKeyPair(lci.RSA, 4192)
	if err != nil {
		fmt.Println("error generating key")
		return nil, err
	}
	fmt.Println("key generated")
	return &manager, nil
}

// GenerateKeyPair is used to generate a public/private key of a non-specific type
func (im *IpnsManager) GenerateKeyPair(keyType, bits int) error {
	priv, pub, err := lci.GenerateKeyPair(keyType, bits)
	if err != nil {
		return err
	}
	im.PrivateKey = priv
	im.PublicKey = pub
	im.KeyType = keyType
	return nil
}

// GenerateEDKeyPair is used to generate an ED25519 keypair
func (im *IpnsManager) GenerateEDKeyPair(bits int) error {
	priv, pub, err := lci.GenerateKeyPair(lci.Ed25519, bits)
	if err != nil {
		return err
	}
	fmt.Println("private key ", priv)
	fmt.Println("public key ", pub)
	im.PrivateKey = priv
	im.PublicKey = pub
	im.KeyType = lci.Ed25519
	return nil
}

func (im *IpnsManager) CreateEntryWithEmbed(ipfsPath string, eol time.Time) (*pb.IpnsEntry, error) {
	if im.KeyType == lci.Ed25519 {
		// see https://github.com/ipfs/go-ipns/pull/5 for more information
		// basically ed25519 public keys are so small they can embed into the peer id
		return nil, errors.New("no need to embed pk when using ed25519")
	}
	fmt.Println("generating ipns entry and embedding")
	entry, err := ipns.Create(im.PrivateKey, []byte(ipfsPath), 1, eol)
	if err != nil {
		fmt.Println("error generating record ", err)
		return nil, err
	}
	err = ipns.EmbedPublicKey(im.PublicKey, entry)
	if err != nil {
		fmt.Println("error embedding key")
		return nil, err
	}
	return entry, nil
}

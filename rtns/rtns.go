package rtns

/*
IPNS related functinonality for temporal
*/
import (
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
	err := manager.GenerateKeyPair(lci.RSA, 4192)
	if err != nil {
		return nil, err
	}
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
	im.PrivateKey = priv
	im.PublicKey = pub
	im.KeyType = lci.Ed25519
	return nil
}

func (im *IpnsManager) CreateEntryAndEmbedPk(ipfsPath string, eol time.Time) (*pb.IpnsEntry, error) {
	entry, err := ipns.Create(im.PrivateKey, []byte(ipfsPath), 1, eol)
	if err != nil {
		return nil, err
	}
	recordPubKeyByte := entry.GetPubKey()
	recordPubKey, err := lci.UnmarshalEd25519PublicKey(recordPubKeyByte)
	if err != nil {
		return nil, err
	}
	err = ipns.EmbedPublicKey(recordPubKey, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

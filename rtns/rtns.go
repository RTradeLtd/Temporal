package rtns

/*
IPNS related functinonality for temporal
*/
import (
	"time"

	lci "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"

	namesys "github.com/ipfs/go-ipfs/namesys"
	pb "github.com/ipfs/go-ipfs/namesys/pb"
	path "github.com/ipfs/go-ipfs/path"
)

// IpnsManager is used to interface with IPNS
type IpnsManager struct {
	PrivateKey lci.PrivKey
	PublicKey  lci.PubKey
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

// GenerateKeyPair is used to generate a public/private key
func (im *IpnsManager) GenerateKeyPair(keyType, bits int) error {
	priv, pub, err := lci.GenerateKeyPair(keyType, bits)
	if err != nil {
		return err
	}
	im.PrivateKey = priv
	im.PublicKey = pub
	return nil
}

func (im *IpnsManager) CreateRoutedEntryData(ipfsPath string, eol time.Time) (*pb.IpnsEntry, error) {
	pathObject := path.FromString(ipfsPath)
	entry, err := namesys.CreateRoutingEntryData(im.PrivateKey, pathObject, 1, eol)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

/*
func (im *IpnsManager) CreateIPNSEntry(ipfsPath string, eol time.Time) (*pb.IpnsEntry, error) {
	pathByte := []byte(ipfsPath)
	entry, err := ipns.Create(im.PrivateKey, pathByte, 1, eol)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
*/

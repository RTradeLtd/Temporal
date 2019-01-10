package ipns

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	pb "gx/ipfs/QmWPFehHmySCdaGttQ48iwF7M6mBRrGE5GSPWKCuMWqJDR/go-ipns/pb"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	u "gx/ipfs/QmNohiVssaPw3KVLZik59DBVGTSm2dGvYT9eoXt5DQ36Yz/go-ipfs-util"
	pstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
	pstoremem "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore/pstoremem"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	proto "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/proto"
)

func testValidatorCase(t *testing.T, priv ci.PrivKey, kbook pstore.KeyBook, key string, val []byte, eol time.Time, exp error) {
	t.Helper()

	match := func(t *testing.T, err error) {
		t.Helper()
		if err != exp {
			params := fmt.Sprintf("key: %s\neol: %s\n", key, eol)
			if exp == nil {
				t.Fatalf("Unexpected error %s for params %s", err, params)
			} else if err == nil {
				t.Fatalf("Expected error %s but there was no error for params %s", exp, params)
			} else {
				t.Fatalf("Expected error %s but got %s for params %s", exp, err, params)
			}
		}
	}

	testValidatorCaseMatchFunc(t, priv, kbook, key, val, eol, match)
}

func testValidatorCaseMatchFunc(t *testing.T, priv ci.PrivKey, kbook pstore.KeyBook, key string, val []byte, eol time.Time, matchf func(*testing.T, error)) {
	t.Helper()
	validator := Validator{kbook}

	data := val
	if data == nil {
		p := []byte("/ipfs/QmfM2r8seH2GiRaC4esTjeraXEachRt8ZsSeGaWTPLyMoG")
		entry, err := Create(priv, p, 1, eol)
		if err != nil {
			t.Fatal(err)
		}

		data, err = proto.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
	}

	matchf(t, validator.Validate(key, data))
}

func TestValidator(t *testing.T) {
	ts := time.Now()

	priv, id, _ := genKeys(t)
	priv2, id2, _ := genKeys(t)
	kbook := pstoremem.NewPeerstore()
	kbook.AddPubKey(id, priv.GetPublic())
	emptyKbook := pstoremem.NewPeerstore()

	testValidatorCase(t, priv, kbook, "/ipns/"+string(id), nil, ts.Add(time.Hour), nil)
	testValidatorCase(t, priv, kbook, "/ipns/"+string(id), nil, ts.Add(time.Hour*-1), ErrExpiredRecord)
	testValidatorCase(t, priv, kbook, "/ipns/"+string(id), []byte("bad data"), ts.Add(time.Hour), ErrBadRecord)
	testValidatorCase(t, priv, kbook, "/ipns/"+"bad key", nil, ts.Add(time.Hour), ErrKeyFormat)
	testValidatorCase(t, priv, emptyKbook, "/ipns/"+string(id), nil, ts.Add(time.Hour), ErrPublicKeyNotFound)
	testValidatorCase(t, priv2, kbook, "/ipns/"+string(id2), nil, ts.Add(time.Hour), ErrPublicKeyNotFound)
	testValidatorCase(t, priv2, kbook, "/ipns/"+string(id), nil, ts.Add(time.Hour), ErrSignature)
	testValidatorCase(t, priv, kbook, "//"+string(id), nil, ts.Add(time.Hour), ErrInvalidPath)
	testValidatorCase(t, priv, kbook, "/wrong/"+string(id), nil, ts.Add(time.Hour), ErrInvalidPath)
}

func mustMarshal(t *testing.T, entry *pb.IpnsEntry) []byte {
	t.Helper()
	data, err := proto.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEmbeddedPubKeyValidate(t *testing.T) {
	goodeol := time.Now().Add(time.Hour)
	kbook := pstoremem.NewPeerstore()

	pth := []byte("/ipfs/QmfM2r8seH2GiRaC4esTjeraXEachRt8ZsSeGaWTPLyMoG")

	priv, _, ipnsk := genKeys(t)

	entry, err := Create(priv, pth, 1, goodeol)
	if err != nil {
		t.Fatal(err)
	}

	testValidatorCase(t, priv, kbook, ipnsk, mustMarshal(t, entry), goodeol, ErrPublicKeyNotFound)

	pubkb, err := priv.GetPublic().Bytes()
	if err != nil {
		t.Fatal(err)
	}

	entry.PubKey = pubkb
	testValidatorCase(t, priv, kbook, ipnsk, mustMarshal(t, entry), goodeol, nil)

	entry.PubKey = []byte("probably not a public key")
	testValidatorCaseMatchFunc(t, priv, kbook, ipnsk, mustMarshal(t, entry), goodeol, func(t *testing.T, err error) {
		if !strings.Contains(err.Error(), "unmarshaling pubkey in record:") {
			t.Fatal("expected pubkey unmarshaling to fail")
		}
	})

	opriv, _, _ := genKeys(t)
	wrongkeydata, err := opriv.GetPublic().Bytes()
	if err != nil {
		t.Fatal(err)
	}

	entry.PubKey = wrongkeydata
	testValidatorCase(t, priv, kbook, ipnsk, mustMarshal(t, entry), goodeol, ErrPublicKeyMismatch)
}

func TestPeerIDPubKeyValidate(t *testing.T) {
	t.Skip("disabled until libp2p/go-libp2p-crypto#51 is fixed")

	goodeol := time.Now().Add(time.Hour)
	kbook := pstoremem.NewPeerstore()

	pth := []byte("/ipfs/QmfM2r8seH2GiRaC4esTjeraXEachRt8ZsSeGaWTPLyMoG")

	sk, pk, err := ci.GenerateEd25519Key(rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}

	pid, err := peer.IDFromPublicKey(pk)
	if err != nil {
		t.Fatal(err)
	}

	ipnsk := "/ipns/" + string(pid)

	entry, err := Create(sk, pth, 1, goodeol)
	if err != nil {
		t.Fatal(err)
	}

	dataNoKey, err := proto.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	testValidatorCase(t, sk, kbook, ipnsk, dataNoKey, goodeol, nil)
}

func genKeys(t *testing.T) (ci.PrivKey, peer.ID, string) {
	sr := u.NewTimeSeededRand()
	priv, _, err := ci.GenerateKeyPairWithReader(ci.RSA, 1024, sr)
	if err != nil {
		t.Fatal(err)
	}

	// Create entry with expiry in one hour
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	ipnsKey := RecordKey(pid)

	return priv, pid, ipnsKey
}

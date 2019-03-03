package testutil

import (
	"bytes"
	"crypto/rand"
	ic "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	pb "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto/pb"
	"testing"
)

func TestBogusPublicKeyGeneration(t *testing.T) {
	public := RandTestBogusPublicKeyOrFatal(t)
	if public.Type() != pb.KeyType_RSA {
		t.Fatalf("Expected public key to be of type RSA but got %s", pb.KeyType_name[int32(public.Type())])
	}
	if val, _ := public.Raw(); val == nil {
		t.Fatal("Expected raw bytes of public key to not be nil")
	}

	otherPublic, _, _ := ic.GenerateRSAKeyPair(512, rand.Reader)
	if public.Equals(otherPublic) {
		t.Fatal("Expect keys of different length to not be equal")
	}
}

func TestBogusPrivateKeyGeneration(t *testing.T) {
	private := RandTestBogusPrivateKeyOrFatal(t)
	if private.Type() != pb.KeyType_RSA {
		t.Fatalf("Expected private key to be of type RSA but got %s", pb.KeyType_name[int32(private.Type())])
	}

	secret := private.GenSecret()
	signedSecret, _ := private.Sign(secret)
	publicBytes, _ := private.GetPublic().Raw()
	public := TestBogusPublicKey(publicBytes)
	if signed, err := public.Verify(secret, signedSecret); !signed || err != nil {
		t.Fatal("Expected to verify signed message")
	}

	encrypted, err := public.Encrypt(secret)
	if err != nil {
		t.Fatal("Failed to encrypt")
	}

	if decrypted, _ := private.Decrypt(encrypted); !bytes.Equal(decrypted, secret) {
		t.Fatal("Decrypting secret did not correspond to plaintext")
	}

	_, otherPrivate, _ := ic.GenerateRSAKeyPair(512, rand.Reader)
	if private.Equals(otherPrivate) {
		t.Fatal("Expect keys of different length to not be equal")
	}

	if val, _ := private.Raw(); val == nil {
		t.Fatal("Expected raw bytes of private key to not be nil")
	}
}

func TestGenerateRandomBogusIdentity(t *testing.T) {
	identity := RandTestBogusIdentityOrFatal(t)
	if identity.ID() == "" {
		t.Fatal("Expected non-empty identity ID")
	}
	if identity.Address() == nil {
		t.Fatal("Expected non-nil address")
	}
	if identity.PrivateKey() == nil {
		t.Fatal("Expected non-nil private key")
	}
	if identity.PublicKey() == nil {
		t.Fatal("Expected non-nil public key")
	}
}

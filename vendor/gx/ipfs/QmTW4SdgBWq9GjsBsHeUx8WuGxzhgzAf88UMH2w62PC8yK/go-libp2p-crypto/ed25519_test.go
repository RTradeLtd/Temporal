package crypto

import (
	"crypto/rand"
	"testing"

	pb "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto/pb"

	"gx/ipfs/QmW7VUmSvhvSGbYbdsh7uRjhGmsYkc9fL8aJ5CorxxrU5N/go-crypto/ed25519"
)

func TestBasicSignAndVerify(t *testing.T) {
	priv, pub, err := GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("hello! and welcome to some awesome crypto primitives")

	sig, err := priv.Sign(data)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := pub.Verify(data, sig)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("signature didn't match")
	}

	// change data
	data[0] = ^data[0]
	ok, err = pub.Verify(data, sig)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatal("signature matched and shouldn't")
	}
}

func TestSignZero(t *testing.T) {
	priv, pub, err := GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, 0)
	sig, err := priv.Sign(data)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := pub.Verify(data, sig)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature didn't match")
	}
}

func TestMarshalLoop(t *testing.T) {
	priv, pub, err := GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PrivateKey", func(t *testing.T) {
		for name, f := range map[string]func() ([]byte, error){
			"Bytes": priv.Bytes,
			"Marshal": func() ([]byte, error) {
				return MarshalPrivateKey(priv)
			},
			"Redundant": func() ([]byte, error) {
				// See issue #36.
				// Ed25519 private keys used to contain the public key twice.
				// For backwards-compatibility, we need to continue supporting
				// that scenario.
				pbmes := new(pb.PrivateKey)
				pbmes.Type = priv.Type()
				data, err := priv.Raw()
				if err != nil {
					t.Fatal(err)
				}

				pbmes.Data = append(data, data[len(data)-ed25519.PublicKeySize:]...)
				return pbmes.Marshal()
			},
		} {
			t.Run(name, func(t *testing.T) {
				bts, err := f()
				if err != nil {
					t.Fatal(err)
				}

				privNew, err := UnmarshalPrivateKey(bts)
				if err != nil {
					t.Fatal(err)
				}

				if !priv.Equals(privNew) || !privNew.Equals(priv) {
					t.Fatal("keys are not equal")
				}

				msg := []byte("My child, my sister,\nThink of the rapture\nOf living together there!")
				signed, err := privNew.Sign(msg)
				if err != nil {
					t.Fatal(err)
				}

				ok, err := privNew.GetPublic().Verify(msg, signed)
				if err != nil {
					t.Fatal(err)
				}

				if !ok {
					t.Fatal("signature didn't match")
				}
			})
		}
	})

	t.Run("PublicKey", func(t *testing.T) {
		for name, f := range map[string]func() ([]byte, error){
			"Bytes": pub.Bytes,
			"Marshal": func() ([]byte, error) {
				return MarshalPublicKey(pub)
			},
		} {
			t.Run(name, func(t *testing.T) {
				bts, err := f()
				if err != nil {
					t.Fatal(err)
				}
				pubNew, err := UnmarshalPublicKey(bts)
				if err != nil {
					t.Fatal(err)
				}

				if !pub.Equals(pubNew) || !pubNew.Equals(pub) {
					t.Fatal("keys are not equal")
				}
			})
		}
	})
}

func TestUnmarshalErrors(t *testing.T) {
	t.Run("PublicKey", func(t *testing.T) {
		t.Run("Invalid data length", func(t *testing.T) {
			pbmes := &pb.PublicKey{
				Type: pb.KeyType_Ed25519,
				Data: []byte{42},
			}

			data, err := pbmes.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			_, err = UnmarshalPublicKey(data)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	})

	t.Run("PrivateKey", func(t *testing.T) {
		t.Run("Redundant public key mismatch", func(t *testing.T) {
			priv, _, err := GenerateEd25519Key(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}

			pbmes := new(pb.PrivateKey)
			pbmes.Type = priv.Type()
			data, err := priv.Raw()
			if err != nil {
				t.Fatal(err)
			}

			// Append the private key instead of the public key.
			pbmes.Data = append(data, data[:ed25519.PublicKeySize]...)
			b, err := pbmes.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			_, err = UnmarshalPrivateKey(b)
			if err == nil {
				t.Fatal("expected an error")
			}
			if err.Error() != "expected redundant ed25519 public key to be redundant" {
				t.Fatalf("invalid error received: %s", err.Error())
			}
		})

		t.Run("Invalid data length", func(t *testing.T) {
			pbmes := &pb.PrivateKey{
				Type: pb.KeyType_Ed25519,
				Data: []byte{42},
			}

			data, err := pbmes.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			_, err = UnmarshalPrivateKey(data)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	})
}

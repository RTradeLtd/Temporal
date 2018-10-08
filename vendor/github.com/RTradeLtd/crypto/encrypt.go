package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keylen  = 32
	saltlen = 8
)

// EncryptManager handles file encryption and decryption
type EncryptManager struct {
	passphrase []byte
}

// NewEncryptManager creates a new EncryptManager
func NewEncryptManager(passphrase string) *EncryptManager {
	return &EncryptManager{[]byte(passphrase)}
}

// Encrypt encrypts given io.Reader, returning the resultant bytes
func (e *EncryptManager) Encrypt(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errors.New("invalid content provided")
	}

	// generate salt, encrypt password for use as a key for a cipher
	salt := make([]byte, saltlen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(e.passphrase), salt, 4096, keylen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// read original content
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// generate an intialization vector for encryption
	encrypted := make([]byte, aes.BlockSize+len(b))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// encrypt
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], b)

	// attach salt to end of encrypted content
	encrypted = append(encrypted, salt...)

	return encrypted, nil
}

// Decrypt decrypts given io.Reader, returning the decrypted bytes
func (e *EncryptManager) Decrypt(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errors.New("invalid content provided")
	}

	// read raw contents
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// retrieve and remove salt
	salt := raw[len(raw)-saltlen:]
	raw = raw[:len(raw)-saltlen]

	// generate cipher
	key := pbkdf2.Key([]byte(e.passphrase), salt, 4096, keylen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// decrypt contents
	stream := cipher.NewCFBDecrypter(block, raw[:aes.BlockSize])
	decrypted := make([]byte, len(raw)-aes.BlockSize)
	stream.XORKeyStream(decrypted, raw[aes.BlockSize:])

	return decrypted, nil
}

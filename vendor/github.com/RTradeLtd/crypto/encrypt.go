package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// if these settings take too long on your server or workstation feel free to modify
	// however please keep in mind these are the settings that Temporal uses.
	// thus if you want to decrypt a file which was encrypted by our Temporal node, you must
	// ensure that the settings match as foillows: keylen = 32, saltlen = 32, nonceSize = 24
	keylen    = 32
	saltlen   = 32
	nonceSize = 24
)

// Protocol is used to configure encryption/decryption methods
type Protocol string

var (
	// GCM allows for usage of AES256-GCM encryption/decryption
	GCM Protocol = "AES256-GCM"
	// CFB allows for usage of AES256-CFB encryption/decryption
	CFB Protocol = "AES256-CFB"
)

// EncryptManager handles file encryption and decryption
type EncryptManager struct {
	passphrase       []byte
	gcmDecryptParams *GCMDecryptParams
	protocol         Protocol
}

// GCMDecryptParams is used to configure decryption for AES256-GCM
type GCMDecryptParams struct {
	CipherKey string
	Nonce     string
}

// NewEncryptManager creates a new EncryptManager
// Default is CFB
func NewEncryptManager(passphrase string) *EncryptManager {
	return &EncryptManager{
		passphrase: []byte(passphrase),
		protocol:   CFB}
}

// WithGCM is used setup, and return EncryptManager for use with AES256-GCM
// the params are expected to be unencrypted, and in hex encoded string format
func (e *EncryptManager) WithGCM(params *GCMDecryptParams) *EncryptManager {
	// set GCM protocol
	e.protocol = GCM
	// set decryption parameters
	e.gcmDecryptParams = params
	// return
	return e
}

// Encrypt is used to handle encryption of objects
func (e *EncryptManager) Encrypt(r io.Reader) ([]byte, error) {
	var out []byte
	switch e.protocol {
	case GCM:
		encryptedData, nonce, cipherKey, err := e.encryptGCM(r)
		if err != nil {
			return nil, err
		}
		// set encrypted data output
		out = encryptedData
		// set gcm decrypt params
		e.gcmDecryptParams = &GCMDecryptParams{
			CipherKey: hex.EncodeToString(cipherKey),
			Nonce:     hex.EncodeToString(nonce),
		}
	case CFB:
		encryptedData, err := e.encryptCFB(r)
		if err != nil {
			return nil, err
		}
		out = encryptedData
	default:
		return nil, fmt.Errorf("no protocol specified")
	}
	return out, nil
}

//eEncryptGCM encrypts given io.Reader using AES256-GCM
// the resultant encrypted bytes, nonce, and cipher are returned
func (e *EncryptManager) encryptGCM(r io.Reader) ([]byte, []byte, []byte, error) {
	if r == nil {
		return nil, nil, nil, errors.New("invalid content provided")
	}
	// create a 32bit cipher key allowing usage for AES256-GCM
	cipherKeyBytes := make([]byte, 32)
	if _, err := rand.Read(cipherKeyBytes); err != nil {
		return nil, nil, nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, nil, err
	}
	block, err := aes.NewCipher(cipherKeyBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	aesGCM, err := cipher.NewGCMWithNonceSize(block, 24)
	if err != nil {
		return nil, nil, nil, err
	}
	dataToEncrypt, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, nil, err
	}
	return aesGCM.Seal(nil, nonce, dataToEncrypt, nil), nonce, cipherKeyBytes, nil
}

// EncryptCFB encrypts given io.Reader using AES256CFB
// the resultant bytes are returned
func (e *EncryptManager) encryptCFB(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errors.New("invalid content provided")
	}

	// generate salt, encrypt password for use as a key for a cipher
	salt := make([]byte, saltlen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	// using sha512 is safer than sha256, but should also be faster on 64bit platforms
	key := pbkdf2.Key(e.passphrase, salt, 4096, keylen, sha512.New)
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

// RetrieveGCMDecryptionParameters is used to retrieve GCM cipher and nonce
// before returning, the cipher and nonce data are formatted, and encrypted
func (e *EncryptManager) RetrieveGCMDecryptionParameters() ([]byte, error) {
	if e.gcmDecryptParams == nil {
		return nil, errors.New("gcm decryption parameters is empty")
	}
	return e.encryptCFB(
		strings.NewReader(fmt.Sprintf(
			"Nonce:\t%s\nCipherKey:\t%s",
			e.gcmDecryptParams.Nonce, e.gcmDecryptParams.CipherKey)))
}

// Decrypt is used to handle decryption of the io.Reader
func (e *EncryptManager) Decrypt(r io.Reader) ([]byte, error) {
	switch e.protocol {
	case CFB:
		return e.decryptCFB(r)
	case GCM:
		return e.decryptGCM(r)
	case GCM:
		if e.gcmDecryptParams == nil {
			return nil, errors.New("no gcm decryption parameters given")
		}
		return e.decryptGCM(r)
	default:
		return nil, fmt.Errorf("invalid invocation, must be one of\nAES256-GCM: EncryptManager::WithGCM::Decrypt\nAES256-CFB: EncryptManager::WithCFB:Decrypt")
	}
}

// DecryptGCM is used to decrypt the given io.Reader using a specified key and nonce
// the key and nonce are expected to be in the format of hex.EncodeToString
func (e *EncryptManager) decryptGCM(r io.Reader) ([]byte, error) {
	if e.gcmDecryptParams == nil {
		return nil, errors.New("gcm decryption parameters is null")
	}
	// decode the key
	decodedKey, err := hex.DecodeString(e.gcmDecryptParams.CipherKey)
	if err != nil {
		return nil, err
	}
	// decode the nonce
	decodedNonce, err := hex.DecodeString(e.gcmDecryptParams.Nonce)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}
	encryptedData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return aesGCM.Open(nil, decodedNonce, encryptedData, nil)
}

// DecryptCFB decrypts given io.Reader which was encrypted using AES256-CFB
// the resulting decrypt bytes are returned
func (e *EncryptManager) decryptCFB(r io.Reader) ([]byte, error) {
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
	// using sha512 is safer than sha256, but should also be faster on 64bit platforms
	key := pbkdf2.Key(e.passphrase, salt, 4096, keylen, sha512.New)
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

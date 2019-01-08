# 🍽 crypto [![GoDoc](https://godoc.org/github.com/RTradeLtd/crypto?status.svg)](https://godoc.org/github.com/RTradeLtd/crypto) [![Build Status](https://travis-ci.com/RTradeLtd/crypto.svg?branch=master)](https://travis-ci.com/RTradeLtd/crypto) [![codecov](https://codecov.io/gh/RTradeLtd/crypto/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/crypto) [![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/crypto)](https://goreportcard.com/report/github.com/RTradeLtd/crypto)

Package crypto provides object encryption utilities for for [Temporal](https://github.com/RTradeLtd/Temporal), an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases. Designed for use on 64bit systems, usage on 32bit systems will probably be a lot slower than usage on 64bit systems. If you are using this intended to provide offline decryption and/or encryption in conjunction with Temporal, with the intent of using with Temporal's encryption/decryption process with your data, do not alter the constant variables as they are what's used by our systems. IF you do not want to use it in conjunctoni with Temporal, and instead perform the encryption/decryption client-side without ever using it server-side with Temporal, you are free to alter the constants to your liking.

It is also available as a command line application:

```sh
$> go get github.com/RTradeLtd/crypto/cmd/temporal-crypto
```

You can then use the tool by calling `temporal-crypto`:

```sh
$> temporal-crypto help
```

## Usage

### Library - Encryption

1) Instantiate `EncryptManager` providing a passphrase used for both AES256-CFB, and AES256-GCM modes
2) Run `EncryptManager.Encrypt` providing a reader for the data you want to encrypt, while specifying the encryption mode.
    2a) For AES256-GCM specify `AES256-GCM` or `aes256-gcm`
    2b) For AES256-CFB specify `AES256-CFB` or `aes256-cfb`

All encryption modes, provide the encrypted data as a byte slice available via the map key of `encryptedData`

AES256-GCM mode provides the randomly generated nonce and cipherkey in a formatted string which was encrypted using AES256-CFB.
The format for the encrypted nonce and cipherkey are of `Nonce:\t<nonce>\nCipherKey:\t<cipherKey>`
Please note that the AES256-GCM encryption process provides the nonce and cipherkey already hex encoded

### Library - Decryption

It is expected that you either use the previously instantiated `EncryptManager`, or a re-instantiated `EncryptManager` with the same passphrase

### CFB Mode

1) Run `EncryptManager.Decrypt` with a reader for your encrypted data, and a nil params argument

### GCM Mode

1) Decrypt the nonce+cipherkey, parsing them for the nonce, and cipherkey values
2) Run `EncrptManager.Decrypt` with a reader for your encrypted data, and a non-nil params argument

## Encryption Process In Depth

We offer two forms of encryption, using either AES256-CFB or AES256-GCM.

### AES256-CFB

When using AES256-CFB, we use the passphrase provided during initialization of the `EncryptManager` and run it through `PBKDF2+SHA512`key derivation function to derive a secure encryption key based on the password. We use this to generate a 32byte key to utilize AES256.

For The salt, we use the secure `rand.Read` to generate a 32byte salt.

Workflow (Encryption): `NewEncryptManager -> Encrypt`
Workflow (Decryption): `NewEncryptManager -> Decrypt`

### AES256-GCM

As a more secure encryption method, we allow the usage of AES256-GCM. For this, we do not let the user decide the cipherkey, and nonce. Like when using AES256-CFB, we leverage `read.Read` to securely generate a random nonce of 24byte, and cipherkey of 32byte, allowing for usage of AES256. Please note that the nonce selection of 24byte as is non-standard. Standad/default nonce is 12byte

As this is intended to be used by Temporal's API, naturally one may be concerned about what we do with the randomly generated cipherkey and nonce. In order to protect the users data, we take the passphrase supplied when instantiating `EncryptManager` and use that combined with our AES256-CFB encryption mechanism to encrypt the cipherkey, and nonce. The encrypted nonce and cipher are in the format of `Nonce:\t<nonce>\nCipherKey:\t<cipherKey>`.

Worfklow (Encryption): `NewEncryptManager -> Encrypt`
Workflow (Decryption): `NewEncryptManager -> Decrypt`
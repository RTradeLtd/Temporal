package crypto

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func Test_EncryptManager(t *testing.T) {
	m := NewEncryptManager("helloworld")

	// open a sample file
	b, err := os.Open("../README.md")
	if err != nil {
		t.Error(err)
		return
	}
	original, err := ioutil.ReadFile("../README.md")
	if err != nil {
		t.Error(err)
		return
	}

	// encrypt file
	encrypted, err := m.Encrypt(b)
	if err != nil {
		t.Error(err)
		return
	}

	// should not be same as original
	if string(original) == string(encrypted) {
		t.Error("encryption did not work")
	}

	// decrypt encrypted bytes
	decrypted, err := m.Decrypt(bytes.NewReader(encrypted))
	if err != nil {
		t.Error(err)
		return
	}

	// compare with original
	if string(original) != string(decrypted) {
		t.Errorf("decrypt failed:\nENCRYPTED=============\n %s DECRYPTED=========\n %s",
			string(original), string(decrypted))
	}
}

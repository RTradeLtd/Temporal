package mini

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/minio/minio-go"
)

const (
	endpoint    = "127.0.0.1:9000"
	keyID       = "C03T49S17RP0APEZDK6M"
	secret      = "q4I9t2MN/6bAgLkbF6uyS7jtQrXuNARcyrm2vvNA"
	bucket      = "test"
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func TestNewMinioManagerNoSecure(t *testing.T) {
	_, err := newMM(false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewMinioManagerSecure(t *testing.T) {
	_, err := newMM(true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListBuckets(t *testing.T) {
	mm, err := newMM(false)
	if err != nil {
		t.Fatal(err)
	}
	buckets, err := mm.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range buckets {
		fmt.Println(v)
	}
}

func TestPutObject(t *testing.T) {
	mm, err := newMM(false)
	if err != nil {
		t.Fatal(err)
	}
	file, err := generateRandomFile()
	if err != nil {
		t.Fatal(err)
	}
	openedFile, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	objName := randString(10)
	fileStats, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	bytesWritten, err := mm.PutObject(bucket, objName, openedFile, fileStats.Size(), minio.PutObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if bytesWritten != fileStats.Size() {
		t.Fatal(errors.New("improper amount of data written to bucket"))
	}
}
func newMM(secure bool) (*MinioManager, error) {
	return NewMinioManager(endpoint, keyID, secret, secure)
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func generateRandomFile() (string, error) {
	randName := randString(20)
	fp := fmt.Sprintf("/tmp/%s", randName)
	file, err := os.Create(fp)
	if err != nil {
		return "", err
	}
	_, err = file.Write([]byte(randString(100)))
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	return fp, nil
}

package mini

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/minio/minio-go"
)

const (
	endpoint    = "127.0.0.1:9000"
	keyID       = "C03T49S17RP0APEZDK6M"
	secret      = "q4I9t2MN/6bAgLkbF6uyS7jtQrXuNARcyrm2vvNA"
	bucket      = "test"
	letterBytes = "abcdefghijklmnopqrstuvwxyz"
)

func TestNewMinioManagerNoSecure(t *testing.T) {
	_, err := newMM(false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewMinioManagerSecure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (no secure cert provided by default)")
	}

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
	_, err = mm.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutAndGetObject(t *testing.T) {
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

	// create bucket, ignore errors
	mm.MakeBucket(map[string]string{"name": bucket})

	// test
	bytesWritten, err := mm.PutObject(bucket, objName, openedFile, fileStats.Size(), minio.PutObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if bytesWritten != fileStats.Size() {
		t.Fatal(errors.New("improper amount of data written to bucket"))
	}

	_, err = mm.PutObject("fake bucket name", objName, openedFile, fileStats.Size(), minio.PutObjectOptions{})
	if err == nil {
		t.Fatal(err)
	}

	objInfo, err := mm.GetObject(bucket, objName, minio.GetObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = objInfo.Stat()
	if err != nil {
		t.Fatal(err)
	}
	_, err = mm.GetObject("fakse bucket namememem", objName, minio.GetObjectOptions{})
	if err == nil {
		t.Fatal("no encountered when one should've been")
	}
	objInfo, err = mm.GetObject(bucket, "definitely a fake object name", minio.GetObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = objInfo.Stat()
	if err == nil {
		t.Fatal(err)
	}
}

func TestMakeBucket(t *testing.T) {
	mm, err := newMM(false)
	if err != nil {
		t.Fatal(err)
	}
	args := make(map[string]string)
	args["name"] = bucket
	err = mm.MakeBucket(args)
	if err == nil {
		t.Fatal("no error encountered one one should've been")
	}
	args["name"] = randString(23)
	err = mm.MakeBucket(args)
	if err != nil {
		t.Fatal(err)
	}
}

func newMM(secure bool) (*MinioManager, error) {
	return NewMinioManager(endpoint, keyID, secret, secure)
}

func randString(n int) string {
	ru := utils.GenerateRandomUtils()
	return ru.GenerateString(n, letterBytes)
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

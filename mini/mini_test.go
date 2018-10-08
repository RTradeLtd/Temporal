package mini

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
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
	// set up
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
	fileStats, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	// create bucket, ignore errors
	mm.MakeBucket(map[string]string{"name": bucket})

	// set some test vars
	var (
		normalObj    = randString(10)
		encryptedObj = randString(10)
		passphrase   = randString(10)
	)

	type args struct {
		object     string
		bucket     string
		passphrase string
	}

	putTests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{normalObj, bucket, ""}, false},
		{"encrypted", args{encryptedObj, bucket, passphrase}, false},
		{"no bucket", args{normalObj, "wut", ""}, true},
	}
	for _, tt := range putTests {
		t.Run(tt.name, func(t *testing.T) {
			var bytesWritten int64
			if bytesWritten, err = mm.PutObject(tt.args.object, openedFile,
				fileStats.Size(), PutObjectOptions{
					Bucket:            tt.args.bucket,
					EncryptPassphrase: tt.args.passphrase,
				}); (err != nil) != tt.wantErr {
				t.Errorf("PutObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// only compare length if not encrypted - if encrypted, addition bytes
			// are added to beginning and end
			if !tt.wantErr && tt.args.passphrase == "" && bytesWritten != fileStats.Size() {
				t.Fatal(errors.New("improper amount of data written to bucket"))
			}
		})
	}

	getTests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{normalObj, bucket, ""}, false},
		{"no bucket", args{normalObj, "wut", ""}, true},
		{"non-existent object", args{"asdf", bucket, ""}, true},
	}
	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err = mm.GetObject(tt.args.object, GetObjectOptions{
				Bucket: tt.args.bucket,
			}); (err != nil) != tt.wantErr {
				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
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

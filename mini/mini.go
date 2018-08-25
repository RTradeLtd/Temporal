package mini

import (
	"errors"
	"io"

	minio "github.com/minio/minio-go"
)

/*
Mini is Temporal's interface with our Minio object storage backend
It provides helper methods to temporal store file ojects that are uploaded to Temporal.
Once objects have been stored in Minio, it can then be uploaded to IPFS

TODO: Add in encryption module that mkaes use of minio server side encryption, allowing a user provided pasword
*/

var DefaultBucketLocation = "us-east-1"

// MinioManager is our helper methods to interface with minio
type MinioManager struct {
	Client *minio.Client
}

// NewMinioManager is used to generate our MinioManager helper struct
func NewMinioManager(endpoint, accessKeyID, secretAccessKey string, secure bool) (*MinioManager, error) {
	mm := &MinioManager{}
	client, err := minio.New(endpoint, accessKeyID, secretAccessKey, !secure)
	if err != nil {
		return nil, err
	}
	_, err = client.ListBuckets()
	if err != nil {
		return nil, err
	}
	mm.Client = client
	return mm, nil
}

// ListBuckets is used to list all known buckets
func (mm *MinioManager) ListBuckets() ([]minio.BucketInfo, error) {
	return mm.Client.ListBuckets()
}

// MakeBucket is a wrapper for te minio MakeBucket method
func (mm *MinioManager) MakeBucket(args map[string]string) error {
	var name, location string
	_, ok := args["name"]
	if !ok {
		return errors.New("name item is missing from args")
	}
	_, ok = args["location"]
	if ok {
		location = args["location"]
	} else {
		location = DefaultBucketLocation
	}
	name = args["name"]
	exists, err := mm.CheckIfBucketExists(name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("bucket already exists")
	}
	return mm.Client.MakeBucket(name, location)
}

// PutObject is a wrapper for the minio PutObject method, returning the number of bytes put or an error
func (mm *MinioManager) PutObject(bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (int64, error) {
	bucketExists, err := mm.CheckIfBucketExists(bucketName)
	if err != nil {
		return 0, err
	}
	if !bucketExists {
		return 0, errors.New("bucket does not exist")
	}
	return mm.Client.PutObject(bucketName, objectName, reader, objectSize, opts)
}

// GetObject is a wrapper for the minio GetObject method
func (mm *MinioManager) GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	exists, err := mm.CheckIfBucketExists(bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("bucket does not exist")
	}
	return mm.Client.GetObject(bucketName, objectName, opts)
}

// RemoveObject is used to remove an object from minio
func (mm *MinioManager) RemoveObject(bucketName, objectName string) error {
	exists, err := mm.CheckIfBucketExists(bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("bucket does not exist")
	}
	return mm.Client.RemoveObject(bucketName, objectName)
}

// CheckIfBucketExists is used to check if a bucket exists
func (mm *MinioManager) CheckIfBucketExists(bucketName string) (bool, error) {
	return mm.Client.BucketExists(bucketName)
}

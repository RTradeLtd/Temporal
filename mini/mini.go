package mini

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/RTradeLtd/crypto"
	"github.com/sirupsen/logrus"

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
	logger logrus.FieldLogger
}

// NewMinioManager is used to generate our MinioManager helper struct
func NewMinioManager(endpoint, accessKeyID, secretAccessKey string, secure bool) (*MinioManager, error) {
	client, err := minio.New(endpoint, accessKeyID, secretAccessKey, secure)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate MinioManager: %s", err)
	}

	// check connection
	if _, err = client.ListBuckets(); err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %s", err)
	}

	return &MinioManager{
		Client: client,
	}, nil
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

type PutObjectOptions struct {
	Bucket            string
	EncryptPassphrase string
	minio.PutObjectOptions
}

// PutObject is a wrapper for the minio PutObject method, returning the number of bytes put or an error
func (mm *MinioManager) PutObject(objectName string, reader io.Reader, objectSize int64,
	opts PutObjectOptions) (int64, error) {
	if opts.Bucket == "" {
		return 0, errors.New("no bucket provided")
	}

	bucketExists, err := mm.CheckIfBucketExists(opts.Bucket)
	if err != nil {
		return 0, err
	}
	if !bucketExists {
		return 0, fmt.Errorf("bucket %s does not exist", opts.Bucket)
	}

	// encrypt if requested
	if opts.EncryptPassphrase != "" {
		encrypted, err := crypto.NewEncryptManager(opts.EncryptPassphrase).Encrypt(reader)
		if err != nil {
			return 0, err
		}

		// update data and metadata
		reader = bytes.NewReader(encrypted)
		objectSize = int64(len(encrypted))
	}

	// store object
	return mm.Client.PutObject(opts.Bucket, objectName, reader, objectSize, opts.PutObjectOptions)
}

type GetObjectOptions struct {
	Bucket string
	minio.GetObjectOptions
}

// GetObject is a wrapper for the minio GetObject method
func (mm *MinioManager) GetObject(objectName string, opts GetObjectOptions) (io.Reader, error) {
	if opts.Bucket == "" {
		return nil, errors.New("no bucket provided")
	}

	exists, err := mm.CheckIfBucketExists(opts.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("bucket does not exist")
	}

	// fetch object
	obj, err := mm.Client.GetObject(opts.Bucket, objectName, opts.GetObjectOptions)
	if err != nil {
		return nil, fmt.Errorf("could not get object %s", objectName)
	}

	// check if object is ok
	_, err = obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not get metadata for object %s", objectName)
	}

	return obj, nil
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

package models_test

import (
	"strings"
	"testing"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/config"
)

func TestEncryptedUploads(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := openDatabaseConnection(t, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ecm := models.NewEncryptedUploadManager(db)
	type args struct {
		user    string
		file    string
		network string
		hash    string
	}
	tests := []args{
		args{"user1", "file1", "network1", "hash1"},
		args{"user1", "file2", "public", "hash2"},
	}
	upload1, err := ecm.NewUpload(tests[0].user, tests[0].file, tests[0].network, tests[0].hash)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Delete(upload1)
	if upload1.FileNameUpper != strings.ToUpper(tests[0].file) {
		t.Fatal("file name to upper failed")

	}
	if upload1.FileNameLower != strings.ToLower(tests[0].file) {
		t.Fatal("file name to lower failed")
	}
	uploads, err := ecm.FindUploadsByUser(tests[0].user)
	if err != nil {
		t.Fatal(err)
	}
	if len(*uploads) != 1 {
		t.Fatal("user should only have 1 upload")
	}
	upload2, err := ecm.NewUpload(tests[1].user, tests[1].file, tests[1].network, tests[1].hash)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Delete(upload2)
	if upload2.FileNameUpper != strings.ToUpper(tests[1].file) {
		t.Fatal("file name to upper failed")

	}
	if upload2.FileNameLower != strings.ToLower(tests[1].file) {
		t.Fatal("file name to lower failed")
	}
	uploads, err = ecm.FindUploadsByUser(tests[0].user)
	if err != nil {
		t.Fatal(err)
	}
	if len(*uploads) != 2 {
		t.Fatal("user should have exactly 2 uploads")
	}
}

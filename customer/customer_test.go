package customer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/RTradeLtd/rtfs/v2"
)

var (
	// actual zdpuAnUGSDoNQoHQ2jpjhPePHEvg26mYLsAAGxr4jkzCWUpde
	emptyObjHash = "zdpuAnUGSDoNQoHQ2jpjhPePHEvg26mYLsAAGxr4jkzCWUpde"
	testHash     = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	testObjHash  = "zdpuAvGhHHFzp7hrs4p3nv4tncWQ4tMCNsaGez1BuZpZUYhpJ"
	testIP       = "192.168.1.101:5001"
)

func Test_Customer_Empty_Object(t *testing.T) {
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ipfs, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		"", 5*time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	obj := Object{
		UploadedRefs:      make(map[string]bool),
		UploadedRootNodes: make(map[string]bool),
	}
	marshaled, err := json.Marshal(&obj)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ipfs.DagPut(marshaled, "json", "cbor")
	if err != nil {
		t.Fatal(err)
	}
	if resp != emptyObjHash {
		t.Fatal("failed to get correct empty object hash")
	}
	manager := NewManager(models.NewUserManager(db.DB), ipfs)
	// defer updating the customer object hash
	// to the default for easy future testing
	defer func() {
		if err := manager.um.UpdateCustomerObjectHash("testuser", emptyObjHash); err != nil {
			t.Fatal(err)
		}
	}()
	size, err := manager.GetDeduplicatedStorageSpaceInBytes("testuser", testHash)
	if err != nil {
		t.Fatal(err)
	}
	if size != 6171 {
		t.Fatal("failed to get size for empty customer object check")
	}
	hash, err := manager.Update("testuser", testHash)
	if hash != testObjHash {
		t.Fatal("failed to get correct object hash")
	}
	user, err := manager.um.FindByUserName("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if user.CustomerObjectHash != testObjHash {
		t.Fatal("failed to set correct customer object hash")
	}
	// now test size calculation for the same test hash
	// this should result in a size of 0 being returned
	size, err = manager.GetDeduplicatedStorageSpaceInBytes("testuser", testHash)
	if err != nil {
		t.Fatal(err)
	}
	if size != 0 {
		t.Fatal("failed to get size for empty customer object check")
	}
	hash, err = manager.Update("testuser", testHash)
	if hash != "" {
		t.Fatal("hash should be empty")
	}
	// create a duplicated linked hash
	unixFSObject, err := ipfs.NewObject("")
	if err != nil {
		t.Fatal(err)
	}
	newHash, err := ipfs.PatchLink(testHash, "hello", unixFSObject, true)
	if err != nil {
		t.Fatal(err)
	}
	if newHash != "QmRYBsa1UiDXfdozyDhbzXhj7PivyjCsBJybYNQ5bBbTBg" {
		t.Fatal("failed to create new hash")
	}
	size, err = manager.GetDeduplicatedStorageSpaceInBytes("testuser", newHash)
	if err != nil {
		t.Fatal(err)
	}
	if size != 2 {
		t.Fatal("failed to calculate correct size")
	}
	newObjHash, err := manager.Update("testuser", newHash)
	if newObjHash != "zdpuAnjTAEDkXQi2aXPkPtscmxRZZQcZr3rw7a2YLvg3pveU8" {
		t.Fatal("failed to properly construct new object hash")
	}
	// now repeat the same test ensuring we get a 0 for size used
	// since we have already stored this hash
	size, err = manager.GetDeduplicatedStorageSpaceInBytes("testuser", newHash)
	if err != nil {
		t.Fatal(err)
	}
	if size != 0 {
		t.Fatal("failed to calculate correct size")
	}
	newObjHash, err = manager.Update("testuser", newHash)
	if newObjHash != "" {
		t.Fatal("failed to properly construct new object hash")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*database.Manager, error) {
	return database.New(cfg, database.Options{SSLModeDisable: true})
}

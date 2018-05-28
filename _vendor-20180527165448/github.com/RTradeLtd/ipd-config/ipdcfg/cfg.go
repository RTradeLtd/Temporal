package ipdcfg

import (
	"encoding/json"
	"fmt"
	"log"
)

// LoadConfig is used to load a config from the passed in ipd object
// and marshal it into a byte slice, suitable for unmarshaling into the supported struct
func (c *Config) LoadConfig(cid string) []byte {
	var m = make(map[string]interface{})
	err := c.Shell.DagGet(cid, &m)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// SaveConfig is used to save a configuration, returning it's content hash
// TODO: TEST
func (c *Config) SaveConfig(m map[string]interface{}) string {
	cid, err := c.Shell.DagPut(m, "json", "cbor")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("The identifier for your dag is ", cid)
	return cid
}

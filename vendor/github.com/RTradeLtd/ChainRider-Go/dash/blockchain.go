package dash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (c *Client) GetBlockchainSyncStatus() (*BlockchainDataSyncStatusResponse, error) {
	url := fmt.Sprintf("%s/sync?token=%s", c.URL, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := BlockchainDataSyncStatusResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

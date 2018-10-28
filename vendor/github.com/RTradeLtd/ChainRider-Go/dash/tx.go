package dash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TransactionByHash is used to get transaction information for a particular hash
func (c *Client) TransactionByHash(txHash string) (*TransactionByHashResponse, error) {
	url := fmt.Sprintf("%s/tx/%s?token=%s", c.URL, txHash, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := TransactionByHashResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// TransactionsForAddress is used to get a list of several transactions for a particular address
func (c *Client) TransactionsForAddress(address string) (*TransactionsForAddressResponse, error) {
	url := fmt.Sprintf("%s/txs?address=%s&token=%s", c.URL, address, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := TransactionsForAddressResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

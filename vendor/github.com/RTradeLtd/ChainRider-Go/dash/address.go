package dash

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// BalanceForAddress returns the balance for an address in duffs
func (c *Client) BalanceForAddress(address string) (int, error) {
	url := fmt.Sprintf("%s/addr/%s/balance?token=%s", c.URL, address, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return 0, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	responseString := string(bodyBytes)
	duffs, err := strconv.ParseInt(responseString, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(duffs), nil
}

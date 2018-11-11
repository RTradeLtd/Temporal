package dash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// PaymentForwardOpts is used to configure payment forward creation
type PaymentForwardOpts struct {
	// DestinationAddress is the address to which dash will be forwarded to
	DestinationAddress string `json:"destination_address"`
}

// CreatePaymentForward is used to create a payment forward
func (c *Client) CreatePaymentForward(opts *PaymentForwardOpts) (*CreatePaymentForwardResponse, error) {
	url := fmt.Sprintf("%s/paymentforward", c.URL)
	payloadString := fmt.Sprintf(
		"{\n  \"destination_address\": \"%s\",\n  \"token\": \"%s\"\n}",
		opts.DestinationAddress, c.Token,
	)
	fmt.Println("url ", url)
	fmt.Println("payload ", payloadString)
	req, err := http.NewRequest("POST", url, strings.NewReader(payloadString))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Println("Headers ", req.Header)
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := CreatePaymentForwardResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// GetPaymentForwardByID is used to lookup a particular payment forward
func (c *Client) GetPaymentForwardByID(id string) (*GetPaymentForwardByIDResponse, error) {
	url := fmt.Sprintf("%s/paymentforward/%s?token=%s", c.URL, id, c.Token)
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
	intf := GetPaymentForwardByIDResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// DeletePaymentForwardByID is used to delete a particular payment forward rule
func (c *Client) DeletePaymentForwardByID(id string) error {
	url := fmt.Sprintf("%s/paymentforward/%s?token=%s", c.URL, id, c.Token)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status %v but received %v", http.StatusOK, resp.StatusCode)
	}
	return nil
}

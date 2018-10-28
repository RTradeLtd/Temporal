package dash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// NewClient is used to initialize our ChainRider client
func NewClient(opts *ConfigOpts) (*Client, error) {
	if opts == nil {
		opts = &ConfigOpts{
			APIVersion:      defaultAPIVersion,
			DigitalCurrency: defaultDigitalCurrency,
			Blockchain:      defaultBlockchain,
			Token:           "test",
		}
	}
	urlFormatted := fmt.Sprintf(urlTemplate, opts.APIVersion, opts.DigitalCurrency, opts.Blockchain)
	c := &Client{
		URL:   urlFormatted,
		Token: opts.Token,
		HC:    &http.Client{},
	}
	// generate our payload
	c.GeneratePayload()
	return c, nil
}

// GeneratePayload is used to generate our payload
func (c *Client) GeneratePayload() {
	c.Payload = fmt.Sprintf(payloadTemplate, c.Token)
}

// GetRateLimit is used to get our rate limit information for the current token
func (c *Client) GetRateLimit() (*RateLimitResponse, error) {
	req, err := http.NewRequest("POST", rateLimitURL, strings.NewReader(c.Payload))
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
	intf := RateLimitResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// GetInformation is used to retrieve  general blockchain information
func (c *Client) GetInformation() (*InformationResponse, error) {
	url := fmt.Sprintf("%s/status?q=getInfo&token=%s", c.URL, c.Token)
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
	intf := InformationResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

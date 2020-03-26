package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	tickerURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
	pricer    *priceChecker
)

type coinPrice struct {
	price       float64
	nextRefresh time.Time
}
type priceChecker struct {
	coins map[string]coinPrice
	mux   *sync.RWMutex
}

// USD contains USD prices for a cryptocurrency
type USD struct {
	Price            float64   `json:"price"`
	Volume24H        float64   `json:"volume_24h"`
	PercentChange1H  float64   `json:"percent_change_1h"`
	PercentChange24H float64   `json:"percent_change_24h"`
	PercentChange7D  float64   `json:"percent_change_7d"`
	MarketCap        float64   `json:"market_cap"`
	LastUpdated      time.Time `json:"last_updated"`
}

func init() {
	pricer = &priceChecker{
		coins: make(map[string]coinPrice),
		mux:   &sync.RWMutex{},
	}
}

// RetrieveUsdPrice is used to retrieve the USD price for a coin from CMC.
//
// Whenever we have a "fresh" coin price that is newer than 10 minutes
// we will return that price instead of querying coinmarketcap. In the event
// of a "stale" value we will hit the coinmarketcap api. If that errors
// then we return both the error, and whatever price we have in-memory
func RetrieveUsdPrice(coin, apiKey string) (float64, error) {
	pricer.mux.RLock()
	if pricer.coins[coin].price != 0 {
		if time.Now().After(pricer.coins[coin].nextRefresh) {
			goto REFRESH
		}
		cost := pricer.coins[coin].price
		pricer.mux.RUnlock()
		return cost, nil
	}
REFRESH:
	pricer.mux.RUnlock()
	pricer.mux.Lock()
	defer pricer.mux.Unlock()
	if pricer.coins[coin].price != 0 && !time.Now().After(pricer.coins[coin].nextRefresh) {
		return pricer.coins[coin].price, nil
	}
	req, err := http.NewRequest("GET", tickerURL, nil)
	if err != nil {
		return pricer.coins[coin].price, err
	}
	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
	q := url.Values{}
	q.Add("slug", coin)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return pricer.coins[coin].price, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return pricer.coins[coin].price, err
	}
	var decode map[string]map[string]interface{}
	if err = json.Unmarshal(body, &decode); err != nil {
		return pricer.coins[coin].price, err
	}
	// we're only interested in the "data" field
	data := decode["data"]
	var (
		datamap, quotemap, usdmap map[string]interface{}
		usd                       USD
		parsed                    bool
	)
	for k := range data {
		out := data[k]
		b, err := json.Marshal(out)
		if err != nil {
			return pricer.coins[coin].price, err
		}
		if err := json.Unmarshal(b, &datamap); err != nil {
			return pricer.coins[coin].price, err
		}
		b, err = json.Marshal(datamap)
		if err != nil {
			return pricer.coins[coin].price, err
		}
		if err := json.Unmarshal(b, &quotemap); err != nil {
			return pricer.coins[coin].price, err
		}
		if quotemap["quote"] != nil {
			b, err = json.Marshal(quotemap["quote"])
			if err := json.Unmarshal(b, &usdmap); err != nil {
				return pricer.coins[coin].price, err
			}
			b, err = json.Marshal(usdmap["USD"])
			if err != nil {
				return pricer.coins[coin].price, err
			}
			if err := json.Unmarshal(b, &usd); err != nil {
				return pricer.coins[coin].price, err
			}
			if usd.Price != 0 {
				parsed = true
				pricer.coins[coin] = coinPrice{
					price:       usd.Price,
					nextRefresh: time.Now().Add(time.Minute * 10),
				}

			}
		}
	}
	if !parsed {
		return pricer.coins[coin].price, errors.New("failed to get price")
	}
	return pricer.coins[coin].price, nil
}

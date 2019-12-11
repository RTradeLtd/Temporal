package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	tickerURL = "https://api.coinmarketcap.com/v1/ticker"
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

// Response used to hold response data from cmc
type Response struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	Rank               string `json:"rank"`
	PriceUsd           string `json:"price_usd"`
	PriceBtc           string `json:"price_btc"`
	TwentyFourHrVolume string `json:"24h_volume_usd"`
	MarketCapUsd       string `json:"market_cap_usd"`
	AvailableSupply    string `json:"available_supply"`
	TotalSupply        string `json:"total_supply"`
	MaxSupply          string `json:"null"`
	PercentChange1h    string `json:"percent_change_1h"`
	PercentChange24h   string `json:"percent_change_24h"`
	PercentChange7d    string `json:"percent_change_7d"`
	LastUpdate         string `json:"last_updated"`
}

func init() {
	pricer = &priceChecker{
		coins: make(map[string]coinPrice),
		mux:   &sync.RWMutex{},
	}
}

// RetrieveUsdPrice is used to retrieve the USD price for a coin from CMC
func RetrieveUsdPrice(coin string) (float64, error) {
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
	url := fmt.Sprintf("%s/%s", tickerURL, coin)
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	var decode []Response
	if err = json.Unmarshal(body, &decode); err != nil {
		return 0, err
	}
	cost, err := strconv.ParseFloat(decode[0].PriceUsd, 64)
	if err != nil {
		return 0, err
	}
	pricer.coins[coin] = coinPrice{
		price:       cost,
		nextRefresh: time.Now().Add(time.Minute * 10),
	}
	return cost, nil
}

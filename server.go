package main

import (
	"encoding/json"
	"github.com/Toorop/go-bittrex"
	"github.com/go-martini/martini"
	"http"
	"ioutil"
	"strconv"
)

func main() {
	bittrex := bittrex.New("fake", "api key")

	type coinbasePrice struct {
		amount   float32
		currency string
	}

	m := martini.Classic()

	m.Group("/prices", func(r martini.Router) {
		m.Get("/BTC", func() (int, string) {
			apiResponse, apiErr := http.Get("https://coinbase.com/api/v1/prices/spot_rate?currency=USD") // TODO support other currencies
			if apiErr != nil {
				return 502, "NETWORK FAIL: " + apiErr // 502 Bad Gateway - upstream failure
			}
			defer apiResponse.Body.Close()
			body, ioErr := ioutil.ReadAll(apiResponse.Body)
			if ioErr != nil {
				return 500, "IO FAIL: " + ioErr
			}
			var price coinbasePrice
			jsonErr := json.Unmarshal(body, &price)
			if jsonErr != nil {
				return 500, "JSON FAIL: " + jsonErr
			}
			return strconv.Itoa(price.amount)
		})

		m.Get("/:coin", func(params martini.Params) (int, string) {
			if params["coin"] == "BTC" {
				return 500, "Tried to handle BTC request like a non-BTC request!"
			}
			ticker, apiErr := bittrex.GetTicker("BTC-" + params["coin"])
			if apiErr != nil {
				return 502, "NETWORK FAIL: " + apiErr
			}
			return 200, strconv.Itoa(ticker.Last)
		})
	})

	m.Run()
}

package main

import (
	"encoding/json"
	"github.com/Toorop/go-bittrex"
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"strconv"
)

func main() {
	bittrex := bittrex.New("fake", "api key")

	type coinbasePrice struct {
		amount   float64
		currency string
	}

	m := martini.Classic()

	m.Group("/prices", func(r martini.Router) {
		m.Get("/BTC", func() (int, string) {
			apiResponse, apiErr := http.Get("https://coinbase.com/api/v1/prices/spot_rate?currency=USD") // TODO support other currencies
			if apiErr != nil {
				return 502, "NETWORK FAIL: " + apiErr.Error() // 502 Bad Gateway - upstream failure
			}
			defer apiResponse.Body.Close()
			body, ioErr := ioutil.ReadAll(apiResponse.Body)
			if ioErr != nil {
				return 500, "IO FAIL: " + ioErr.Error()
			}
			var priceObject interface{}
			jsonErr := json.Unmarshal(body, &priceObject)
			if jsonErr != nil {
				return 500, "JSON FAIL: " + jsonErr.Error()
			}
			priceObjectMap := priceObject.(map[string]interface{})
			price := priceObjectMap["amount"].(string)
			return 200, "Price of BTC is $" + price
		})

		m.Get("/:coin", func(params martini.Params) (int, string) {
			ticker, apiErr := bittrex.GetTicker("BTC-" + params["coin"])
			if apiErr != nil {
				return 502, "NETWORK FAIL: " + apiErr.Error()
			}
			return 200, "Price of " + params["coin"] + " is " + strconv.FormatFloat(ticker.Last, 'f', -1, 64) + " BTC"
		})
	})

	m.Run()
}

package main

import (
	"encoding/json"
	"github.com/Toorop/go-bittrex"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"net/http"
	"strconv"
)

func stringifyFiatValue(rawValue float64) string {
	return strconv.FormatFloat(rawValue, 'f', 2, 64)
}

func stringifyBTCValue(rawValue float64) string {
	return strconv.FormatFloat(rawValue, 'f', 8, 64)
}

func getBTCPrice() (result float64, err error) {
	apiResponse, err := http.Get("https://coinbase.com/api/v1/prices/spot_rate?currency=USD") // TODO support other currencies
	if err != nil {
		return
	}
	defer apiResponse.Body.Close()
	body, err := ioutil.ReadAll(apiResponse.Body)
	if err != nil {
		return
	}
	var priceObject interface{}
	err = json.Unmarshal(body, &priceObject)
	if err != nil {
		return
	}
	priceObjectMap := priceObject.(map[string]interface{})
	price := priceObjectMap["amount"].(string)
	return strconv.ParseFloat(price, 64)
}

func getPrice(coin string) (result float64, err error) {
	if coin == "BTC" {
		return 1, nil
	} else {
		bittrex := bittrex.New("fake", "api key") // TODO come up with a non-lame way to do this

		ticker, err := bittrex.GetTicker("BTC-" + coin)
		return ticker.Last, err
	}
}

type rawCoin struct {
	Name     string
	Abbr     string
	Quantity float64
}

type fullCoin struct {
	Name     string
	Abbr     string
	Quantity string
	BtcValue string
	UsdValue string
}

type templateData struct {
	BtcNetWorth string
	UsdNetWorth string
	CoinValues  []fullCoin
}

func makeTemplateData(coins []rawCoin) templateData {
	result := templateData{}
	btcNetWorth := 0.0
	usdNetWorth := 0.0
	for _, coin := range coins {
		resultCoin := fullCoin{}
		resultCoin.Name = coin.Name
		resultCoin.Abbr = coin.Abbr
		resultCoin.Quantity = stringifyBTCValue(coin.Quantity)
		btcPrice, _ := getPrice(coin.Abbr)
		btcValue := btcPrice * coin.Quantity
		resultCoin.BtcValue = stringifyBTCValue(btcValue)
		priceOfBTC, _ := getBTCPrice()
		usdValue := btcValue * priceOfBTC
		resultCoin.UsdValue = stringifyFiatValue(usdValue)
		btcNetWorth += btcValue
		usdNetWorth += usdValue
		result.CoinValues = append(result.CoinValues, resultCoin)
	}
	result.BtcNetWorth = stringifyBTCValue(btcNetWorth)
	result.UsdNetWorth = stringifyFiatValue(usdNetWorth)
	return result
}

func loadData() []rawCoin {
	filename := "data.json"
	rawData, _ := ioutil.ReadFile(filename) // TODO don't discard this error
	var result []rawCoin
	json.Unmarshal(rawData, &result) // TODO or this error
	return result
}

func main() {
	type coinbasePrice struct {
		amount   float64
		currency string
	}

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Group("/prices", func(r martini.Router) {
		r.Get("/BTC", func() (int, string) {
			price, err := getBTCPrice()
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Price of BTC is $" + stringifyFiatValue(price)
		})

		r.Get("/:coin", func(params martini.Params) (int, string) {
			price, err := getPrice(params["coin"])
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Price of " + params["coin"] + " is " + stringifyBTCValue(price)
		})
	})

	m.Group("/value", func(r martini.Router) {
		r.Get("/BTC/:count", func(params martini.Params) (int, string) {
			count, err := strconv.ParseFloat(params["count"], 64)
			price, err := getBTCPrice()
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Value of " + params["count"] + " BTC is $" + stringifyFiatValue(price*count)
		})

		r.Get("/:coin/:count", func(params martini.Params) (int, string) {
			count, err := strconv.ParseFloat(params["count"], 64)
			price, err := getPrice(params["coin"])
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Value of " + params["count"] + " " + params["coin"] + " is " + stringifyBTCValue(price*count) + " BTC"
		})
	})

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", makeTemplateData(loadData()))
	})

	m.Run()
}

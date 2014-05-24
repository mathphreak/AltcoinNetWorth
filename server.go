package main

import (
	"encoding/json"
	"github.com/Toorop/go-bittrex"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
)

func round(rawValue float64, decimalPlaces int) float64 {
	scale := math.Pow10(decimalPlaces)
	return math.Trunc(rawValue*scale) / scale // TODO don't naively use floor
}

func roundFiatValue(rawValue float64) float64 {
	return round(rawValue, 2)
}

func roundBTCValue(rawValue float64) float64 {
	return round(rawValue, 8)
}

func stringify(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
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
	Quantity float64
	BtcValue float64
	UsdValue float64
}

type templateData struct {
	BtcNetWorth float64
	UsdNetWorth float64
	CoinValues  []fullCoin
}

func makeTemplateData(coins []rawCoin) templateData {
	result := templateData{}
	for _, coin := range coins {
		resultCoin := fullCoin{}
		resultCoin.Name = coin.Name
		resultCoin.Abbr = coin.Abbr
		resultCoin.Quantity = coin.Quantity
		btcPrice, _ := getPrice(coin.Abbr)
		btcValue := btcPrice * coin.Quantity
		resultCoin.BtcValue = roundBTCValue(btcValue)
		priceOfBTC, _ := getBTCPrice()
		usdValue := btcValue * priceOfBTC
		resultCoin.UsdValue = roundFiatValue(usdValue)
		result.BtcNetWorth += btcValue
		result.UsdNetWorth += usdValue
		result.CoinValues = append(result.CoinValues, resultCoin)
	}
	result.BtcNetWorth = roundBTCValue(result.BtcNetWorth)
	result.UsdNetWorth = roundFiatValue(result.UsdNetWorth)
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
			return 200, "Price of BTC is $" + stringify(roundFiatValue(price))
		})

		r.Get("/:coin", func(params martini.Params) (int, string) {
			price, err := getPrice(params["coin"])
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Price of " + params["coin"] + " is " + stringify(roundBTCValue(price))
		})
	})

	m.Group("/value", func(r martini.Router) {
		r.Get("/BTC/:count", func(params martini.Params) (int, string) {
			count, err := strconv.ParseFloat(params["count"], 64)
			price, err := getBTCPrice()
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Value of " + params["count"] + " BTC is $" + stringify(roundFiatValue(price*count))
		})

		r.Get("/:coin/:count", func(params martini.Params) (int, string) {
			count, err := strconv.ParseFloat(params["count"], 64)
			price, err := getPrice(params["coin"])
			if err != nil {
				return 500, "ERROR: " + err.Error()
			}
			return 200, "Value of " + params["count"] + " " + params["coin"] + " is " + stringify(roundBTCValue(price*count)) + " BTC"
		})
	})

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", makeTemplateData([]rawCoin{rawCoin{"Bitcoin", "BTC", 0.01512}, rawCoin{"Dogecoin", "DOGE", 115023}}))
	})

	m.Run()
}

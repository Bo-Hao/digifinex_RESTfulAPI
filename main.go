package main

import (
	"fmt"
)

const (
	dfxRestURI = "https://openapi.digifinex.com/v3"
)

const (
	appKey    = "acc8a49e0fcc4b"
	appSecret = "5fb78e8084f0536c286471b0f2a5def7f7a556a03c"
)

func main() {
	app := DigifinexClient{
		appKey:    appKey,
		appSecret: appSecret,
	}

	//fmt.Println(app.PlaceLimitOrder("btc_usdt", "spot", "buy", "8000", "1"))
	//fmt.Println(app.GetBalances())
	fmt.Println(app.GetOpenOrders())

	/* response, err := app.doRequest("POST", "/spot/order/new", map[string]interface{}{
		"symbol":    "btc_usdt",
		"type":      "buy",
		"amount":    1,
		"price":     "8100",
		"post_only": 0,
	}, true)
	if err != nil {
		panic(err)
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}
	fmt.Println(string(textRes)) */
}

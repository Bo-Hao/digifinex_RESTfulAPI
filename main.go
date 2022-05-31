package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	dfxRestURI = "https://openapi.digifinex.com/v3"
)

const (
	apiKey    = "acc8a49e0fcc4b"
	apiSecret = "5fb78e8084f0536c286471b0f2a5def7f7a556a03c"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// ws
	/* o := SpotLocalOrderbook(ctx, "ETH_BTC")

	for i := 0; i < 10; i++ {
		fmt.Println(o.GetAsks())
		fmt.Println(o.GetBids())
		time.Sleep(time.Second)
	} */

	/* app := DigifinexClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	} */
	app := DigifinexClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}

	go app.TradeReportWebsocket(ctx, []string{"ETH_BTC"})

	//fmt.Println(app.PlaceLimitOrder("btc_usdt", "spot", "buy", "8000", "1"))
	//fmt.Println(app.GetBalances())
	//fmt.Println(app.GetOpenOrders())

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

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Ctrl + C pressed!")
	cancel()
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	//jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

//var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (client *DigifinexClient) PlaceLimitOrder(symbol, product, side, price, qty string) (string, error) {

	response, err := client.doRequest("POST", "/"+strings.ToLower(product)+"/order/new", map[string]interface{}{
		"symbol":    strings.ToLower(symbol), // base_quote
		"type":      strings.ToLower(side),   // buy or sell
		"amount":    qty,
		"price":     price,
		"post_only": 0,
	}, true)

	if err != nil {
		return "", err
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var resMap placeOrderRes
	json.Unmarshal(textRes, &resMap)

	if resMap.Code != 0 {
		return "", errors.New(fmt.Sprint(resMap.Code))
	}

	orderId := resMap.OrderID
	return orderId, err
}

func (client *DigifinexClient) PlacePostOnlyOrder(symbol, product, side, price, qty string) (string, error) {
	response, err := client.doRequest("POST", "/"+strings.ToLower(product)+"/order/new", map[string]interface{}{
		"symbol":    strings.ToLower(symbol), // base_quote
		"type":      strings.ToLower(side),   // buy or sell
		"amount":    qty,
		"price":     price,
		"post_only": 1,
	}, true)

	if err != nil {
		return "", err
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var resMap placeOrderRes
	json.Unmarshal(textRes, &resMap)

	if resMap.Code != 0 {
		return "", errors.New(fmt.Sprint(resMap.Code))
	}

	orderId := resMap.OrderID
	return orderId, err
}

func (client *DigifinexClient) PlaceMarketOrder(symbol, product, side, qty string) (string, error) {
	buffer := bytes.NewBufferString(strings.ToLower(side))
	buffer.WriteString("_market")

	response, err := client.doRequest("POST", "/"+strings.ToLower(product)+"/order/new", map[string]interface{}{
		"symbol":    strings.ToLower(symbol), // base_quote
		"type":      buffer.String(),         // buy_market or sell_market
		"amount":    qty,
		"post_only": 0,
	}, true)

	if err != nil {
		return "", err
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var resMap placeOrderRes
	json.Unmarshal(textRes, &resMap)

	if resMap.Code != 0 {
		return "", errors.New(fmt.Sprint(resMap.Code))
	}

	orderId := resMap.OrderID
	return orderId, err
}

func (client *DigifinexClient) CancelOrder(product string, orderIds []string) (successIds []string, err error) {
	response, err := client.doRequest("POST", "/"+strings.ToLower(product)+"/order/cancel", map[string]interface{}{
		"order_id": strings.Join(orderIds, ","),
	}, true)

	if err != nil {
		return successIds, err
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return successIds, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var res cancelOrderRes
	json.Unmarshal(textRes, &res)

	if res.Code != 0 {
		return successIds, errors.New(fmt.Sprint(res.Code))
	}

	successIds = res.Success
	return successIds, err
}

func (client *DigifinexClient) GetBalances() ([][]string, bool) {
	response, err := client.doRequest("GET", "/spot/assets", map[string]interface{}{}, true)
	if err != nil {
		return [][]string{}, false
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return [][]string{}, false
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var res getBalanceRes
	json.Unmarshal(textRes, &res)

	if res.Code != 0 {
		log.Println(errors.New(fmt.Sprint(res.Code)))
		return [][]string{}, false
	}

	var balances [][]string
	for _, v := range res.List {
		balances = append(balances, []string{v.Currency, fmt.Sprint(v.Free), fmt.Sprint(v.Total)})
	}

	return balances, true
}

func (client *DigifinexClient) GetOpenOrders() ([][]string, bool) {
	// []string{oid, symbol, product, subaccount, price, qty, side, orderType4, UnfilledQty}
	response, err := client.doRequest("GET", "/spot/order/current", map[string]interface{}{}, true)
	if err != nil {
		return [][]string{}, false
	}
	textRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return [][]string{}, false
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		panic(fmt.Errorf(string(textRes)))
	}

	var res getOpenOrdersRes
	json.Unmarshal(textRes, &res)

	if res.Code != 0 {
		log.Println(errors.New(fmt.Sprint(res.Code)))
		return [][]string{}, false
	}

	var openOrders [][]string
	// []string{oid, symbol, product, subaccount, price, qty, side, orderType, UnfilledQty}
	for _, v := range res.Data {
		qty, _ := decimal.NewFromString(v.Amount)
		filled, _ := decimal.NewFromString(v.ExecutedAmount)

		openOrders = append(openOrders, []string{v.OrderID, v.Symbol, "spot", "", v.Price, v.Amount, v.Type, "limit", qty.Sub(filled).String()})
	}

	return openOrders, true
}

/* func (client *DigifinexClient) GetTradeReports() ([][]string, bool) {
	// []string{oid, symbol, product, subaccount, price, qty, side, orderType4, fee, filledQty, timestamp, isMaker}
} */

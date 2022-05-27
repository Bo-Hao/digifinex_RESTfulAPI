package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type tradeRes struct {
	Method string      `json:"method"`
	Params []openOrder `json:"params"`
	ID     interface{} `json:"id"`
}

type openOrder struct {
	Amount    string `json:"amount"`
	Filled    string `json:"filled"`
	ID        string `json:"id"`
	Mode      int    `json:"mode"`
	Notional  string `json:"notional"`
	Price     string `json:"price"`
	PriceAvg  string `json:"price_avg"`
	Side      string `json:"side"`
	Status    int    `json:"status"`
	Symbol    string `json:"symbol"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
}

// trade report
func (client *DigifinexClient) TradeReportWebsocket(ctx context.Context, marketList []string) {
	var url string = "wss://openapi.digifinex.com/ws/v1/"
	client.updateOnErr(false)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		client.updateOnErr(true)
	}

	authSubMsg, err := authSubscribeMsg(client.apiKey, client.apiSecret)
	if err != nil {
		client.updateOnErr(true)
	}

	err = conn.WriteMessage(websocket.TextMessage, authSubMsg)
	if err != nil {
		client.updateOnErr(true)
		client.ws.conn.SetReadDeadline(time.Now().Add(time.Second))
	}

	_, authMsg, err := conn.ReadMessage()
	if err != nil {
		client.updateOnErr(true)
	}

	authRes, _ := decodingMap(authMsg)
	if authRes["error"] == nil {
		client.updateOnErr(true)
	}

	tradeSubMsg, err := tradeReportSubscribeMessage(marketList)
	if err != nil {
		client.updateOnErr(true)
	}

	err = conn.WriteMessage(websocket.TextMessage, tradeSubMsg)
	if err != nil {
		client.updateOnErr(true)
		client.ws.conn.SetReadDeadline(time.Now().Add(time.Second))
	}

	_, tradeMsg, err := conn.ReadMessage()
	if err != nil {
		client.updateOnErr(true)
	}

	tradeRes, _ := decodingMap(tradeMsg)
	if tradeRes["error"] == nil {
		client.updateOnErr(true)
	}

	client.ws.conn = conn

	// mainloop
mainloop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if client.ws.conn == nil {
				break mainloop
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err, string(msg))
				time.Sleep(time.Millisecond * 500)
				break mainloop
			}

			errh := client.handleTradeReportMsg(msg)
			if errh != nil {
				log.Println(errh)
				break mainloop
			}
		} // end select
		// if there is something wrong that the WS should be reconnected.
		if client.checkOnErr() {
			break mainloop
		}

		time.Sleep(time.Millisecond)
	} // end for

	conn.Close()
	client.ws.conn.Close()

	// if it is manual work.
	if !client.checkOnErr() {
		return
	}

	client.TmpBranch.Lock()

	client.openOrderBranch.RLock()
	openOrders := client.openOrderBranch.openOrders
	client.openOrderBranch.RUnlock()

	client.TmpBranch.openOrders = openOrders
	client.TmpBranch.Unlock()

	client.TradeReportWebsocket(ctx, marketList)
}

func authSubscribeMsg(apikey, apisecret string) ([]byte, error) {
	nonce := strconv.FormatInt(time.Now().UnixMilli(), 10)
	key := []byte(apisecret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(nonce))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	params := make(map[string]interface{})
	params["id"] = 0
	params["method"] = "server.auth"
	params["params"] = []string{apikey, nonce, signature}

	req, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// provide private subscribtion message.
func tradeReportSubscribeMessage(marketList []string) ([]byte, error) {
	// prepare authentication message.
	param := make(map[string]interface{})
	param["id"] = 0
	param["method"] = "order.subscribe"
	param["params"] = marketList

	req, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (client *DigifinexClient) handleTradeReportMsg(msg []byte) error {
	msgMap, err := decodingMap(msg)
	if err != nil {
	}

	if _, ok := msgMap["method"]; ok && msgMap["method"] == "order.update" {
		buffer := bytes.NewBuffer(msg)
		var res tradeRes
		reader, err0 := zlib.NewReader(buffer)
		if err0 != nil {

		}
		defer reader.Close()
		content, err1 := ioutil.ReadAll(reader)
		if err1 != nil {

		}
		json.Unmarshal(content, &res)

		// test the snapshot (?)
		client.updateOrders(res)
	}

	return nil
}

func (client *DigifinexClient) updateOrders(res tradeRes) {
	if res.Method == "order.update" {
		return
	}
	client.openOrderBranch.Lock()
	defer client.openOrderBranch.Unlock()
	for _, v := range res.Params {
		if _, ok := client.openOrderBranch.openOrders[v.ID]; !ok {
			client.openOrderBranch.openOrders[v.ID] = v
		}

		switch v.Status {
		case 0: // unfilled
			client.openOrderBranch.openOrders[v.ID] = v // reassign make sure there is order in map. (?)

		case 1: // partially filled
			filled, err := decimal.NewFromString(v.Filled)
			if err != nil {
				panic(err)
			}
			oldFilled, err := decimal.NewFromString(client.openOrderBranch.openOrders[v.ID].Filled)
			if err != nil {
				panic(err)
			}

			executeQty := filled.Sub(oldFilled)
			if executeQty.Equal(decimal.Zero) {
				executeQty = filled
			}

			isMaker := "false"
			if v.Type == "limit" {
				isMaker = "true"
			}
			//[]string{oid, symbol, product, subaccount, price, qty, side, orderType4, fee, filledQty, timestamp, isMaker}
			tradeReport := []string{v.ID, v.Symbol, "spot", "", v.PriceAvg, v.Amount, v.Side, v.Type, "0", executeQty.String(), fmt.Sprint(v.Timestamp), isMaker}
			client.dumpTradeReport(tradeReport)

		case 2: // fully filled
			filled, err := decimal.NewFromString(v.Filled)
			if err != nil {
				panic(err)
			}
			oldFilled, err := decimal.NewFromString(client.openOrderBranch.openOrders[v.ID].Filled)
			if err != nil {
				panic(err)
			}

			executeQty := filled.Sub(oldFilled)
			if executeQty.Equal(decimal.Zero) {
				executeQty = filled
			}

			isMaker := "false"
			if v.Type == "limit" {
				isMaker = "true"
			}
			//[]string{oid, symbol, product, subaccount, price, qty, side, orderType4, fee, filledQty, timestamp, isMaker}
			tradeReport := []string{v.ID, v.Symbol, "spot", "", v.PriceAvg, v.Amount, v.Side, v.Type, "0", executeQty.String(), fmt.Sprint(v.Timestamp), isMaker}
			client.dumpTradeReport(tradeReport)

			delete(client.openOrderBranch.openOrders, v.ID)
		case 3: // canceled unfilled
			delete(client.openOrderBranch.openOrders, v.ID)

		case 4: // canceled partially filled
			delete(client.openOrderBranch.openOrders, v.ID)
		}

	}
}

func (client *DigifinexClient) dumpTradeReport(tradeReport []string) {
	client.TradeReportBranch.Lock()
	defer client.TradeReportBranch.Unlock()
	client.TradeReportBranch.TradeReports = append(client.TradeReportBranch.TradeReports, tradeReport)
}

func (client *DigifinexClient) updateOnErr(on bool) {
	client.onErrBranch.Lock()
	defer client.onErrBranch.Unlock()
	client.onErrBranch.onErr = on
}

func (client *DigifinexClient) checkOnErr() bool {
	client.onErrBranch.RLock()
	defer client.onErrBranch.RUnlock()
	return client.onErrBranch.onErr
}

/* func (Mc *MaxClient) parseTradeReportSnapshotMsg(msgMap map[string]interface{}) error {
	jsonbody, _ := json.Marshal(msgMap["t"])
	var newTrades []Trade
	json.Unmarshal(jsonbody, &newTrades)
	Mc.trackingTradeReports(newTrades)

	return nil
}

func (Mc *MaxClient) parseTradeReportUpdateMsg(msgMap map[string]interface{}) error {
	jsonbody, _ := json.Marshal(msgMap["t"])
	var newTrades []Trade
	json.Unmarshal(jsonbody, &newTrades)
	Mc.tradeReportsArrived(newTrades)

	return nil
}

func (Mc *MaxClient) trackingTradeReports(snapshottrades []Trade) error {
	Mc.WsClient.TmpBranch.Lock()
	oldTrades := Mc.WsClient.TmpBranch.Trades
	Mc.WsClient.TmpBranch.Trades = []Trade{}
	Mc.WsClient.TmpBranch.Unlock()

	if len(oldTrades) == 0 {
		Mc.UpdateTrades(snapshottrades)
		return nil
	}

	untrades := Mc.ReadUnhedgeTrades()

	tradeMap := map[int64]struct{}{}
	oldTrades = append(oldTrades, untrades...)
	for i := 0; i < len(oldTrades); i++ {
		tradeMap[oldTrades[i].Id] = struct{}{}
	}

	untracked, tracked := make([]Trade, 0, 130), make([]Trade, 0, 130)
	for i := 0; i < len(snapshottrades); i++ {
		if _, ok := tradeMap[snapshottrades[i].Id]; !ok {
			untracked = append(untracked, snapshottrades[i])
		} else {
			tracked = append(tracked, snapshottrades[i])
		}
	}

	Mc.UpdateTrades(tracked)
	if len(untracked) > 0 {
		fmt.Println("trade snapshot:", untracked)
		Mc.TradesArrived(untracked)
		Mc.tradeReportsArrived(untracked)
	}

	return nil
}

func (Mc *MaxClient) tradeReportsArrived(trades []Trade) {
	Mc.TradeReportBranch.Lock()
	Mc.TradeReportBranch.TradeReports = append(Mc.TradeReportBranch.TradeReports, trades...)
	Mc.TradeReportBranch.Unlock()
	Mc.TradesArrived(trades)
} */

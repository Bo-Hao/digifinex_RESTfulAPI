package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/Bo-Hao/mapbook"
	"github.com/gorilla/websocket"
)

type OrderbookBranch struct {
	Market     string
	bookBranch struct {
		bids mapbook.BidBook
		asks mapbook.AskBook
		sync.RWMutex
	}

	onErrBranch struct {
		onErr bool
		sync.RWMutex
	}

	ws struct {
		conn *websocket.Conn
	}

	pingpongBranch struct {
		lastPingPongTime time.Time
		sync.RWMutex
	}
}

func SpotLocalOrderbook(ctx context.Context, symbol string) *OrderbookBranch {
	var o OrderbookBranch
	o.Market = symbol
	o.bookBranch.asks = *mapbook.NewAskBook(false)
	o.bookBranch.bids = *mapbook.NewBidBook(false)
	go o.maintain(ctx, symbol)
	return &o
}

// trade report
func (o *OrderbookBranch) maintain(ctx context.Context, symbol string) {
	var url string = "wss://openapi.digifinex.com/ws/v1/"
	o.updateErr(false)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		o.updateErr(true)
	}

	subMsg, err := orderbookSubscribeMsg(symbol)
	if err != nil {
		o.updateErr(true)
	}

	err = conn.WriteMessage(websocket.TextMessage, subMsg)
	if err != nil {
		o.ws.conn.SetReadDeadline(time.Now().Add(time.Second))
		o.updateErr(true)
	}

	o.ws.conn = conn
	defer conn.Close()

	// ping pong
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := o.sendPingPong()
				if err != nil {
					log.Println("ping pong failed")
				}

				o.pingpongBranch.RLock()
				lastPong := o.pingpongBranch.lastPingPongTime
				o.pingpongBranch.RUnlock()
				if time.Since(lastPong).Seconds() > 180 {
					o.updateErr(true)
				}

			}
			if o.checkErr() {
				return
			}
		}
	}()

	// mainloop
mainloop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if o.ws.conn == nil {
				o.updateErr(true)
				break mainloop
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err, string(msg))
				time.Sleep(time.Millisecond * 500)
				o.updateErr(true)
				break mainloop
			}

			errh := o.handleOrderbookMsg(msg)
			if errh != nil {
				log.Println(errh)
				o.updateErr(true)
				break mainloop
			}
		} // end select
		if o.checkErr() {
			break mainloop
		}
		time.Sleep(time.Millisecond)
	} // end for

	if !o.checkErr() {
		return
	}

	o.maintain(ctx, symbol)
}

// provide private subscribtion message.
func orderbookSubscribeMsg(symbol string) ([]byte, error) {
	// making signature
	// prepare authentication message.
	param := make(map[string]interface{})
	param["id"] = 0
	param["method"] = "depth.subscribe"
	param["params"] = []string{symbol}

	req, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodingMap(message []byte) (res map[string]interface{}, err error) {
	// decode by zlib
	buffer := bytes.NewBuffer(message)
	reader, err0 := zlib.NewReader(buffer)
	if err0 != nil {
		return res, err0
	}
	defer reader.Close()
	content, err1 := ioutil.ReadAll(reader)
	if err1 != nil {
		return res, err1
	}
	err = json.Unmarshal(content, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OrderbookBranch) handleOrderbookMsg(msg []byte) (err error) {
	type res struct {
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
		ID     interface{}   `json:"id"`
	}
	msgMap, err := decodingMap(msg)
	if err != nil {
		log.Println("fail to decode")
	}

	if _, ok := msgMap["error"]; ok {
		if msgMap["result"].(string) == "pong" {
			o.pingpongBranch.Lock()
			o.pingpongBranch.lastPingPongTime = time.Now()
			o.pingpongBranch.Unlock()
		}
		return nil
	}

	if u, ok := msgMap["method"]; ok && u == "depth.update" {
		b, _ := json.Marshal(msgMap)
		var r res
		json.Unmarshal(b, &r)
		snapshot := r.Params[0].(bool)

		n := r.Params[1].(map[string]interface{})
		askByte, _ := json.Marshal(n["asks"])
		bidByte, _ := json.Marshal(n["bids"])

		var askStr, bidStr [][]string
		json.Unmarshal(askByte, &askStr)
		json.Unmarshal(bidByte, &bidStr)

		if snapshot {
			o.bookBranch.asks.Snapshot(askStr)
			o.bookBranch.bids.Snapshot(bidStr)
		} else {
			o.bookBranch.asks.Update(askStr)
			o.bookBranch.bids.Update(bidStr)
		}
	}

	return nil
}

func (o *OrderbookBranch) updateErr(on bool) {
	o.onErrBranch.Lock()
	defer o.onErrBranch.Unlock()
	o.onErrBranch.onErr = on
}

func (o *OrderbookBranch) checkErr() bool {
	o.onErrBranch.RLock()
	defer o.onErrBranch.RUnlock()
	on := o.onErrBranch.onErr
	return on
}

func (o *OrderbookBranch) sendPingPong() error {
	ping := make(map[string]interface{})
	ping["id"] = 0
	ping["method"] = "server.ping"
	ping["params"] = []string{}

	message, err := json.Marshal(ping)
	if err != nil {
		return err
	}
	if err := o.ws.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		o.ws.conn.SetReadDeadline(time.Now().Add(time.Second))
		return err
	}
	return nil
}

func (o *OrderbookBranch) GetBids() ([][]string, bool) {
	return o.bookBranch.bids.GetAll()
}

func (o *OrderbookBranch) GetAsks() ([][]string, bool) {
	return o.bookBranch.asks.GetAll()
}

func (o *OrderbookBranch) Close() {
	o.updateErr(true)
	o.ws.conn.Close()
}

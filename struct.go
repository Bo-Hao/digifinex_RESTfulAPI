package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type DigifinexClient struct {
	apiKey         string
	apiSecret      string
	deafultTimeout time.Duration

	ws WSClient

	onErrBranch struct {
		onErr bool
		sync.RWMutex
	}

	openOrderBranch struct {
		openOrders map[string]openOrder // order id
		sync.RWMutex
	}

	TradeReportBranch struct {
		TradeReports [][]string
		sync.RWMutex
	}

	TmpBranch struct {
		openOrders map[string]openOrder // order id
		sync.RWMutex
	}
}

type WSClient struct {
	conn *websocket.Conn
}

func parseToString(val interface{}) string {
	switch t := val.(type) {
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", t)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", t)
	case float32, float64:
		return fmt.Sprintf("%.8f", t)
	case string:
		return t
	default:
		panic(fmt.Errorf("invalid value type", t))
	}
}

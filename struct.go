package main

import (
	"fmt"
	"time"
)

type DigifinexClient struct {
	appKey         string
	appSecret      string
	deafultTimeout time.Duration
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

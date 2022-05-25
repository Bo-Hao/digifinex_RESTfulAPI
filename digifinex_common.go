package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func hmacSha256(key []byte, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}

func (client *DigifinexClient) doRequest(method string, path string, params map[string]interface{}, sign bool) (*http.Response, error) {
	inputParam := url.Values{}
	for key, val := range params {
		inputParam.Add(key, parseToString(val))
	}
	encodedParams := inputParam.Encode()

	var (
		req *http.Request
		err error
	)
	if method == "GET" {
		if encodedParams != "" {
			req, err = http.NewRequest(method, dfxRestURI+path+"?"+encodedParams, nil)
		} else {
			req, err = http.NewRequest(method, dfxRestURI+path, nil)
		}
	} else {
		req, err = http.NewRequest(method, dfxRestURI+path, strings.NewReader(encodedParams))
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create request " + err.Error())
	}
	// header
	req.Header.Add("ACCESS-TIMESTAMP", strconv.Itoa(int(time.Now().Unix())))
	if sign {
		req.Header.Add("ACCESS-KEY", client.appKey)
		req.Header.Add("ACCESS-SIGN", hmacSha256([]byte(client.appSecret), []byte(encodedParams)))
		//fmt.Println(req.Header.Get("ACCESS-SIGN"))
	}
	fmt.Println(req.Header)
	c := &http.Client{
		Transport: &http.Transport{},
		Timeout:   client.deafultTimeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"ginWeb/utils/exchange"
	"net/http"
	"sync"
	"time"
)

const (
	spot = "https://api3.binance.com"
	swap = "https://fapi.binance.com"
)

// Binance 币安
type Binance struct {
	ApiKey string
	Secret string
	Window int
}

// Sign 生成签名
func (receiver Binance) Sign(msg string) (string, error) {
	if receiver.Secret == "" || receiver.ApiKey == "" {
		err := errors.New("secret or api key is empty")
		return "", err
	}
	h := hmac.New(sha256.New, []byte(receiver.Secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Request 同步请求
func (receiver Binance) Request(item exchange.ExInterface) {
	req, err := http.NewRequest(item.Method(), item.Route(), nil)
	if err != nil {
		item.SetResult(nil, err)
		return
	}
	client := http.Client{Timeout: time.Duration(receiver.Window) * time.Millisecond}
	res, err := client.Do(req)
	if err != nil {
		item.SetResult(nil, err)
		return
	}
	var data []byte
	_, err = res.Body.Read(data)
	if res.StatusCode != 200 {
		item.SetResult(nil, errors.New(fmt.Sprintf("http status code %d: %s", res.StatusCode, string(data))))
		return
	}

	item.SetResult(data, err)
	return
}

func (receiver Binance) AsyncRequests(reqs ...exchange.ExInterface) error {
	if len(reqs) == 0 {
		return errors.New("reqs is empty")
	}

	// 一次请求
	if len(reqs) == 1 {
		receiver.Request(reqs[0])
		return nil
	}
	wg := sync.WaitGroup{}

	for _, req := range reqs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			receiver.Request(req)
		}()
	}
	wg.Wait()

	return nil
}

func (receiver Binance) SyncRequests(reqs ...exchange.ExInterface) error {
	if len(reqs) == 0 {
		return errors.New("reqs is empty")
	}
	for _, req := range reqs {
		receiver.Request(req)
	}
	return nil
}

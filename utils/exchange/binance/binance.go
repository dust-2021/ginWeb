package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"ginWeb/utils/exchange"
)

const (
	spot = "https://api3.binance.com"
	swap = "https://fapi.binance.com"
)

// Binance 币安
type Binance struct {
	apiKey string
	secret string
}

func (receiver Binance) Sign(msg string) (data string, err error) {
	if receiver.secret == "" || receiver.apiKey == "" {
		err = errors.New("secret or api key is empty")
	}
	h := hmac.New(sha256.New, []byte(receiver.secret))
	h.Write([]byte(msg))
	return
}

func request(item *exchange.ExInterface) (data *exchange.ExResp) {
	return
}

func (receiver Binance) AsyncRequest(reqs ...*exchange.ExInterface) (responses *[]exchange.ExResp, err error) {
	if len(reqs) == 0 {
		err = errors.New("reqs is empty")
		return
	}

	return
}

func (receiver Binance) SyncRequest(reqs ...*exchange.ExInterface) (responses *[]exchange.ExResp, err error) {
	return
}

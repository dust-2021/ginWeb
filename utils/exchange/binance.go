package exchange

import (
	"crypto/hmac"
	"crypto/sha256"
)

const (
	spot = "https://api3.binance.com"
	swap = "https://fapi.binance.com"
)

// Binance TODO binance接口
type Binance struct {
	apiKey string
	secret string
}

func (receiver Binance) Sign(msg string) (data string, err error) {
	h := hmac.New(sha256.New, []byte(receiver.secret))
	h.Write([]byte(msg))
	return
}

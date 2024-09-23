package binanceApi

import (
	"encoding/json"
	"errors"
	"ginWeb/utils/exchange/binance"
	"github.com/shopspring/decimal"
)

type SwapPrice struct {
	Data   map[string]interface{}
	result struct {
		Data []byte
		Err  error
	}
}

type swapPriceResp []map[string]interface{}

func (s *SwapPrice) Route() string {
	return binance.Swap + "/fapi/v1/ticker/price"
}

func (s *SwapPrice) Method() string {
	return "GET"
}

func (s *SwapPrice) Sign() bool {
	return false
}

func (s *SwapPrice) ReqData() map[string]interface{} {
	return s.Data
}

func (s *SwapPrice) SetResult(resp []byte, err error) {
	s.result.Data = resp
	s.result.Err = err
}

func (s *SwapPrice) GetResult() (map[string]decimal.Decimal, error) {
	var resp swapPriceResp
	err := json.Unmarshal(s.result.Data, &resp)
	if err != nil {
		return nil, err
	}
	result := make(map[string]decimal.Decimal, len(resp))
	for _, item := range resp {
		key, f1 := item["symbol"]
		value, f2 := item["price"]
		if !f1 || !f2 {
			return nil, errors.New("binanceApi get result error")
		}
		revalue, err := decimal.NewFromString(value.(string))
		if err != nil {
			return nil, err
		}
		sKey := key.(string)
		result[sKey] = revalue
	}
	return result, nil
}

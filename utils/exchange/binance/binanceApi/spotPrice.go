package binanceApi

import (
	"encoding/json"
	"errors"
	"ginWeb/utils/exchange/binance"
	"github.com/shopspring/decimal"
)

// SpotPrice 现货价格
type SpotPrice struct {
	Data   map[string]interface{}
	result struct {
		Data  []byte
		Error error
	}
}

type spotPriceResp []map[string]interface{}

func (s *SpotPrice) Route() string {
	return binance.Spot + "/api/v3/ticker/price"
}

func (s *SpotPrice) Method() string {
	return "GET"
}

func (s *SpotPrice) Sign() bool {
	return false
}

func (s *SpotPrice) ReqData() map[string]interface{} {
	return s.Data
}

func (s *SpotPrice) SetResult(resp []byte, err error) {
	s.result.Data = resp
	s.result.Error = err
}

func (s *SpotPrice) GetResult() (map[string]decimal.Decimal, error) {
	if s.result.Error != nil {
		return nil, s.result.Error
	}
	var resp spotPriceResp
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

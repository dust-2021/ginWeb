package binanceApi

import (
	"encoding/json"
	"errors"
	"ginWeb/utils/exchange/binance"
)

type KeyCheck struct {
	Data   map[string]interface{}
	result struct {
		Data  []byte
		Error error
	}
}

type data struct {
	EnableFutures              bool `json:"enableFutures"`
	EnableSpotAndMarginTrading bool `json:"enableSpotAndMarginTrading"`
	PermitsUniversalTransfer   bool `json:"permitsUniversalTransfer"`
}

func (receiver *KeyCheck) Route() string {
	return binance.Spot + "/sapi/v1/account/apiRestrictions"
}

func (receiver *KeyCheck) Method() string {
	return "GET"
}

func (receiver *KeyCheck) Sign() bool {
	return true
}

func (receiver *KeyCheck) ReqData() map[string]interface{} {
	return receiver.Data
}

func (receiver *KeyCheck) SetResult(resp []byte, err error) {
	receiver.result.Data = resp
	receiver.result.Error = err
}

func (receiver *KeyCheck) GetResult() (bool, error) {

	if receiver.result.Error != nil {
		return false, receiver.result.Error
	}
	var data data
	err := json.Unmarshal(receiver.result.Data, &data)
	if err != nil {
		return false, err
	}
	if !data.EnableFutures {
		return false, errors.New("lack enableFutures")
	}
	if !data.EnableSpotAndMarginTrading {
		return false, errors.New("lack enableSpotAndMarginTrading")
	}
	if !data.PermitsUniversalTransfer {
		return false, errors.New("lack permitsUniversalTransfer")
	}
	return true, nil
}

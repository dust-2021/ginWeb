package binanceApi

import (
	"encoding/json"
	"errors"
	"ginWeb/utils/exchange/binance"
)

// KeyCheck 检查key权限
type KeyCheck struct {
	Data   map[string]interface{}
	result struct {
		Data  []byte
		Error error
	}
}

type keyCheckResp struct {
	EnableFutures              bool `json:"enableFutures"`
	EnableSpotAndMarginTrading bool `json:"enableSpotAndMarginTrading"`
	PermitsUniversalTransfer   bool `json:"permitsUniversalTransfer"`
}

func (k *KeyCheck) Route() string {
	return binance.Spot + "/sapi/v1/account/apiRestrictions"
}

func (k *KeyCheck) Method() string {
	return "GET"
}

func (k *KeyCheck) Sign() bool {
	return true
}

func (k *KeyCheck) ReqData() map[string]interface{} {
	return k.Data
}

func (k *KeyCheck) SetResult(resp []byte, err error) {
	k.result.Data = resp
	k.result.Error = err
}

func (k *KeyCheck) GetResult() (bool, error) {

	if k.result.Error != nil {
		return false, k.result.Error
	}
	var data keyCheckResp
	err := json.Unmarshal(k.result.Data, &data)
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

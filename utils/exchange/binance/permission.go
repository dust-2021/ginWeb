package binance

type Permission struct {
	Data   map[string]interface{}
	result struct {
		Data  []byte
		Error error
	}
}

func (receiver Permission) Route() string {
	return ""
}
func (receiver Permission) Method() string {
	return "GET"
}

func (receiver Permission) Sign() bool {
	return true
}

func (receiver Permission) ReqData() map[string]interface{} {
	return receiver.Data
}

func (receiver Permission) SetResult(resp []byte, err error) {
	receiver.result.Data = resp
	receiver.result.Error = err
}

func (receiver Permission) GetResult() (bool, error) {
	var data bool
	if receiver.result.Error == nil {
		data = true
	}
	return data, receiver.result.Error
}

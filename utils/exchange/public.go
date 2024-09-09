package exchange

// ExInterface 请求接口类型
type ExInterface interface {
	Method() string
	Route() string
	Data() map[string]interface{}
	Sign() bool
	ID() string
}

// ExResp 接口请求结果
type ExResp struct {
	Id    string
	Data  string
	Error error
}

// Exchange 交易所类
type Exchange interface {
	// Sign 请求签名
	Sign(msg string) (data string, err error)
	// AsyncRequest 异步发送请求
	AsyncRequest(reqs ...*ExInterface) (responses *[]ExResp, err error)
	// SyncRequest 同步发送请求
	SyncRequest(reqs ...*ExInterface) (responses *[]ExResp, err error)
}

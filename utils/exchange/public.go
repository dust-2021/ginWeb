package exchange

// ExInterface 请求接口类型
type ExInterface interface {
	Method() string
	Route() string
	ReqData() map[string]interface{}
	Sign() bool
	SetResult([]byte, error)
}

// ExResp 接口请求结果
type ExResp struct {
	Data  interface{}
	Error error
}

// Exchange 交易所类
type Exchange interface {
	// Request 同步请求
	Request(ExInterface)
	// AsyncRequests 异步发送请求
	AsyncRequests(...ExInterface) error
	// SyncRequests 同步发送请求
	SyncRequests(...ExInterface) error
}

package exchange

/*
封装交易所接口
每个ExInterface代表一次接口访问，使用Exchange对象的request方法后会将结果存入接口访问对象

Example:
	inter := ExInterface{}
	ex := Exchange{}

	// 单次同步访问
	ex.Request(inter)
	// 多次同步访问
	ex.SyncRequest(inter1, inter2, ...)
	// 多次异步访问
	ex.AsyncRequest(inter1, inter2, ...)

	// 获取请求结果，不同接口实现逻辑不同
	inter.GetResult()
*/

// ExInterface 请求接口类型
type ExInterface interface {
	Method() string
	Route() string
	ReqData() map[string]interface{}
	Sign() bool
	SetResult([]byte, error)
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

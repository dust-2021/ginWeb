package exchange

type ExInterface struct {
	Route  string
	Method string
	data   interface{}
}

type Exchange interface {
	Sign(msg string) (data string, err error)
	AsyncRequest(method string, route string, sign bool, data interface{}, ch chan interface{}) (err error)
	SyncRequest(method string, route string, sign bool, data interface{}) (err error)
}

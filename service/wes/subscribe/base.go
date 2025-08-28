package subscribe

import "ginWeb/service/wes"

// Pub 事件订阅接口类
type Pub interface {
	// Subscribe 订阅该事件
	Subscribe(connection *wes.Connection) error
	// UnSubscribe 取消订阅
	UnSubscribe(connection *wes.Connection) error
	// Publish 向收听者发送消息
	Publish([]byte, *wes.Connection) error
	// Start 启动事件
	Start(...interface{}) error
	// Shutdown 关闭事件
	Shutdown() error
}

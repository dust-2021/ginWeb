package subscribe

import "ginWeb/service/wes"

// Pub 事件订阅接口类
type Pub interface {
	// Subscribe 订阅该事件
	Subscribe(connection *wes.Connection, args ...any) error
	// UnSubscribe 取消订阅
	UnSubscribe(connection *wes.Connection, args ...any) error
	// IsSuber 判断是否订阅者
	IsSuber(connection *wes.Connection) bool
	// Publish 向收听者发送消息
	Publish([]byte, *wes.Connection) error
	// Start 启动事件
	Start(...any) error
	// Shutdown 关闭事件
	Shutdown() error
}

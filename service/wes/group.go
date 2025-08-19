package wes

import (
	"fmt"
	"ginWeb/utils/loguru"
)

// 已注册的ws处理逻辑
var tasks = make(map[string]*handler)

// RegisterHandler 注册ws处理函数，key已存则触发panic
func RegisterHandler(key string, f ...handleFunc) {
	if _, flag := tasks[key]; flag {
		loguru.Logger.Fatalf("duplicate register ws handler: %s", key)
		return
	}
	var h handler = f
	tasks[key] = &h

}

// Group ws处理组
type Group struct {
	node    string
	middles []handleFunc
}

var BasicGroup = &Group{node: "", middles: []handleFunc{}}

// Use 添加中间件
func (g *Group) Use(f ...handleFunc) {
	g.middles = append(g.middles, f...)
}

// Group 创建一个子组
func (g *Group) Group(name string, f ...handleFunc) *Group {
	var key string
	if g.node == "" {
		key = name
	} else {
		key = fmt.Sprintf("%s.%s", g.node, name)
	}
	return &Group{
		node:    key,
		middles: append(g.middles, f...),
	}
}

// Register 在组上创建处理函数
func (g *Group) Register(route string, f ...handleFunc) {
	var key string
	if g.node == "" {
		key = route
	} else {
		key = fmt.Sprintf("%s.%s", g.node, route)
	}
	RegisterHandler(key, append(g.middles, f...)...)
}

// NewGroup 在根组上新建组
func NewGroup(name string, f ...handleFunc) *Group {
	return BasicGroup.Group(name, f...)
}

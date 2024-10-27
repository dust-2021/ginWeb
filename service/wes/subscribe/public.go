package subscribe

import (
	"context"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"strings"
	"sync"
	"time"
)

type Pub interface {
	// Subscribe 订阅事件
	Subscribe(connection *wes.Connection)
	// UnSubscribe 取消订阅
	UnSubscribe(connection *wes.Connection)
	// Publish 发送数据
	Publish([]byte)
	// ShutDown 关闭事件
	ShutDown()
}

var Pubs = make(map[string]Pub)
var PubsLock = sync.RWMutex{}

// PeriodPub 周期性订阅事件
type PeriodPub struct {
	Name        string
	subscribers map[*wes.Connection]struct{}
	lock        sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	period time.Duration
	f      func() []byte
}

func (p *PeriodPub) Subscribe(c *wes.Connection) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.subscribers[c] = struct{}{}
}

func (p *PeriodPub) UnSubscribe(c *wes.Connection) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.subscribers, c)
}

func (p *PeriodPub) Publish(v []byte) {
	p.lock.Lock()
	defer p.lock.Unlock()
	var shouldDelete = make([]*wes.Connection, 0)
	for c := range p.subscribers {
		err := c.Send(v)
		if err != nil {
			shouldDelete = append(shouldDelete, c)
		}
	}
	for _, conn := range shouldDelete {
		loguru.SimpleLog(loguru.Debug, "WS", "delete sub from "+conn.RemoteAddr().String())
		delete(p.subscribers, conn)
	}
}

func (p *PeriodPub) ShutDown() {
	PubsLock.Lock()
	defer PubsLock.Unlock()
	p.cancel()
	delete(Pubs, p.Name)
}

func (p *PeriodPub) do() {
	ticker := time.NewTicker(p.period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.Publish(p.f())
		case <-p.ctx.Done():
			return
		}
	}
}

// NewPeriodPub 注册订阅事件，将f函数结果发送至每个订阅者，d为发送周期
func NewPeriodPub(name string, d time.Duration, f func() []byte) *PeriodPub {
	ctx, cancel := context.WithCancel(context.Background())
	pub := &PeriodPub{
		Name:        name,
		subscribers: make(map[*wes.Connection]struct{}),
		lock:        sync.RWMutex{},
		ctx:         ctx,
		cancel:      cancel,
		period:      d,
		f:           f,
	}
	Pubs[name] = pub
	go pub.do()
	return pub
}

func subHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}
	var keys []string
	for _, v := range w.Request.Params {
		name, flag := v.(string)
		if !flag {
			w.Result(dataType.WrongData, fmt.Sprintf("param invalid %v", v))
			return
		}
		keys = append(keys, name)
	}

	PubsLock.RLock()
	defer PubsLock.RUnlock()
	var failedKeys = make([]string, 0)
	for _, name := range keys {
		if p, f := Pubs[name]; f {
			p.Subscribe(w.Conn)
		} else {
			failedKeys = append(failedKeys, name)
		}
	}
	if len(failedKeys) > 0 {
		w.Result(dataType.NotFound, strings.Join(failedKeys, ","))
	}
	w.Result(dataType.Success, "")
}

func unsubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}

	var keys []string
	for _, v := range w.Request.Params {
		name, flag := v.(string)
		if !flag {
			w.Result(dataType.WrongData, fmt.Sprintf("param invalid %v", v))
			return
		}
		keys = append(keys, name)
	}
	PubsLock.RLock()
	defer PubsLock.RUnlock()
	for _, name := range keys {
		pub, ok := Pubs[name]
		if !ok {
			continue
		}
		pub.UnSubscribe(w.Conn)
	}
	w.Result(dataType.Success, "")
}

func init() {
	wes.RegisterHandler("subscribe", subHandle)
	wes.RegisterHandler("unsubscribe", unsubHandle)
}

package subscribe

import (
	"context"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/scheduler"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"github.com/robfig/cron/v3"
	"strings"
	"sync"
	"time"
)

// Publisher 周期性订阅事件
type Publisher struct {
	Name string

	// 订阅事件的ws连接
	subscribers map[*wes.Connection]struct{}
	// ws连接管理锁
	subContainerLock sync.RWMutex

	closed bool

	ctx    context.Context
	cancel context.CancelFunc
	// 生成结果的函数
	f func() []byte
	// cron任务ID，period类型为0
	taskId cron.EntryID
}

func (p *Publisher) Subscribe(c *wes.Connection) {
	p.subContainerLock.Lock()
	defer p.subContainerLock.Unlock()
	p.subscribers[c] = struct{}{}
	loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("user from %s get in channel %s", c.RemoteAddr().String(), p.Name))
}

func (p *Publisher) UnSubscribe(c *wes.Connection) {
	p.subContainerLock.Lock()
	defer p.subContainerLock.Unlock()
	delete(p.subscribers, c)
	loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("user from %s get out channel %s", c.RemoteAddr().String(), p.Name))
}

func (p *Publisher) Publish(v []byte, sender *wes.Connection) {
	p.subContainerLock.Lock()
	defer p.subContainerLock.Unlock()
	if p.closed || v == nil {
		return
	}
	var shouldDelete = make([]*wes.Connection, 0)
	for c := range p.subscribers {
		// 不向发送者发送消息
		if c == sender {
			continue
		}
		err := c.Send(v)
		if err != nil {
			shouldDelete = append(shouldDelete, c)
		}
	}
	for _, conn := range shouldDelete {
		loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("delete publisher \"%s\" from %s", p.Name, conn.RemoteAddr().String()))
		delete(p.subscribers, conn)
	}
}

func (p *Publisher) Shutdown() {
	wes.PubsLock.Lock()
	defer wes.PubsLock.Unlock()
	p.cancel()
	if p.taskId != 0 {
		scheduler.App.Remove(p.taskId)
	}
	p.taskId = 0
	p.closed = true
}

func (p *Publisher) Start(timer string) error {
	p.subContainerLock.Lock()
	defer p.subContainerLock.Unlock()
	p.closed = false
	if timer == "" || p.f == nil {
		return nil
	}
	duration, err := time.ParseDuration(timer)
	if err == nil {
		go p.periodDo(duration)
		return nil
	}
	err = p.cronDo(timer)
	return err
}

func (p *Publisher) periodDo(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.Publish(p.f(), nil)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Publisher) cronDo(cron string) error {
	_, err := scheduler.App.AddFunc(cron, func() {
		p.Publish(p.f(), nil)
	})
	return err
}

// NewPublisher 注册订阅事件，将f函数结果发送至每个订阅者，d为发送周期, d为空字符串时不会注册为定时事件
func NewPublisher(name string, d string, f ...func() []byte) wes.Pub {
	ctx, cancel := context.WithCancel(context.Background())
	pub := &Publisher{
		Name:             name,
		subscribers:      make(map[*wes.Connection]struct{}),
		subContainerLock: sync.RWMutex{},
		ctx:              ctx,
		cancel:           cancel,
		closed:           false,
	}
	if len(f) == 0 {
		pub.f = nil
	} else {
		pub.f = f[0]
	}
	err := pub.Start(d)
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "WS", "failed create pub")
	}
	wes.Pubs[name] = pub
	return pub
}

// ws订阅事件接口
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

	wes.PubsLock.RLock()
	defer wes.PubsLock.RUnlock()
	var failedKeys = make([]string, 0)
	for _, name := range keys {
		if p, f := wes.Pubs[name]; f {
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

// ws取消事件订阅接口
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
	wes.PubsLock.RLock()
	defer wes.PubsLock.RUnlock()
	for _, name := range keys {
		pub, ok := wes.Pubs[name]
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

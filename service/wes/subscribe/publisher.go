package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/service/scheduler"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

// ws已注册订阅事件
var pubs = make(map[string]*Publisher)

// 已注册订阅事件读写锁
var pubsLock = sync.RWMutex{}

// GetPub 查找订阅事件
func GetPub(name string) (*Publisher, bool) {
	pubsLock.RLock()
	defer pubsLock.RUnlock()
	pub, ok := pubs[name]
	return pub, ok
}

// SetPub 设置订阅事件
func SetPub(name string, pub *Publisher) bool {
	pubsLock.Lock()
	defer pubsLock.Unlock()
	if pub == nil {
		delete(pubs, name)
		return true
	}
	if _, ok := pubs[name]; ok {
		return false
	}
	pubs[name] = pub
	return true
}

type resp struct {
	SenderId   int64       `json:"senderId"`
	SenderName string      `json:"senderName"`
	Timestamp  int64       `json:"timestamp"`
	Data       interface{} `json:"data"`
}

type Subscriber struct {
	Conn  *wes.Connection
	Pub   *Publisher
	Muted bool
}

// Publisher 订阅事件
type Publisher struct {
	Name string

	// 订阅事件的ws连接
	subscribers map[*wes.Connection]*Subscriber
	// 对象读写锁
	lock sync.RWMutex
	// 开启状态
	closed bool

	ctx    context.Context
	cancel context.CancelFunc
	// 生成结果的函数
	f func() string
	// cron任务ID，period类型为0
	taskId cron.EntryID
}

func (p *Publisher) Subscribe(c *wes.Connection) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.subscribers[c]; ok {
		return errors.New("already subscribed")
	}
	p.subscribers[c] = &Subscriber{
		Conn:  c,
		Pub:   p,
		Muted: true,
	}
	loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("user from %s subscribe channel %s", c.RemoteAddr().String(), p.Name))
	c.DoneHook(func() {
		_ = p.UnSubscribe(c)
	})
	return nil
}

func (p *Publisher) UnSubscribe(c *wes.Connection) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.subscribers, c)
	loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("user from %s unsubscribe channel %s", c.RemoteAddr().String(), p.Name))
	return nil
}

func (p *Publisher) Publish(v string, sender *wes.Connection) error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.closed {
		return errors.New("publish closed")
	}
	if v == "" {
		return errors.New("msg is empty")
	}

	var senderName = ""
	var senderId int64 = 0
	if sender != nil {
		if sub, ok := p.subscribers[sender]; ok && sub.Muted {
			return errors.New("you have been muted")
		}
		senderId, senderName, _ = sender.UserInfo()
		if senderId == 0 {
			senderName = sender.RemoteAddr().String()
		}
	}
	var r = resp{
		SenderId:   senderId,
		SenderName: senderName,
		Timestamp:  time.Now().UnixMilli(),
		Data:       v,
	}
	data, _ := json.Marshal(r)
	for c := range p.subscribers {
		// 不向发送者发送消息
		if c == sender {
			continue
		}

		_ = c.Send(data)

	}
	return nil
}

func (p *Publisher) Shutdown() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cancel()
	if p.taskId != 0 {
		scheduler.App.Remove(p.taskId)
	}
	clear(p.subscribers)
	p.taskId = 0
	p.closed = true
	return nil
}

func (p *Publisher) Start(timer string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.closed {
		return errors.New("publisher is already start")
	}
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
			err := p.Publish(p.f(), nil)
			if err != nil {
				loguru.SimpleLog(loguru.Error, "publish fail", err.Error())
				return
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Publisher) cronDo(cron string) error {
	_, err := scheduler.App.AddFunc(cron, func() {
		err := p.Publish(p.f(), nil)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "publish fail", err.Error())
			return
		}
	})
	return err
}

// NewPublisher 注册并启动订阅事件，将f函数结果发送至每个订阅者，d为发送周期, d为空字符串时不会注册为定时事件
func NewPublisher(name string, d string, f ...func() string) *Publisher {
	ctx, cancel := context.WithCancel(context.Background())
	pub := &Publisher{
		Name:        name,
		subscribers: make(map[*wes.Connection]*Subscriber),
		lock:        sync.RWMutex{},
		ctx:         ctx,
		cancel:      cancel,
		closed:      true,
	}
	if len(f) == 0 {
		pub.f = nil
	} else {
		pub.f = f[0]
	}
	err := pub.Start(d)
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "WS", "failed start pub:"+err.Error())
	}
	SetPub(name, pub)
	return pub
}

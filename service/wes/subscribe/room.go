package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

var rooms = make(map[string]*Room)

var roomsLock sync.RWMutex

// GetRoom 根据room名称查找room
func GetRoom(name string) (*Room, bool) {
	roomsLock.RLock()
	defer roomsLock.RUnlock()
	r, ok := rooms[name]
	return r, ok
}

// SetRoom 修改room，不存在则创建，如果是新建room则返回true
func SetRoom(name string, room *Room) bool {
	roomsLock.Lock()
	defer roomsLock.Unlock()
	if room == nil {
		delete(rooms, name)
		return true
	}
	if _, ok := rooms[name]; ok {
		return false
	}
	rooms[name] = room
	return true
}

type Room struct {
	uuid  string
	subs  map[*wes.Connection]struct{}
	lock  sync.RWMutex
	owner *wes.Connection

	refreshCtx context.Context
	refresh    context.CancelFunc
	closed     bool
}

func (r *Room) UUID() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.uuid
}

func (r *Room) Owner() (int64, string) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	userId, username, _ := r.owner.UserInfo()
	return userId, username
}

type mateInfo struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Addr  string `json:"addr"`
	Owner bool   `json:"owner"`
}

// Mates 所有成员
func (r *Room) Mates() []mateInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	resp := make([]mateInfo, 0)
	for c := range r.subs {
		id, name, _ := c.UserInfo()
		resp = append(resp, mateInfo{
			Name:  name,
			Id:    strconv.FormatInt(id, 10),
			Addr:  c.RemoteAddr().String(),
			Owner: c == r.owner,
		})
	}
	return resp
}

// 生命周期管理
func (r *Room) closer() {
	for {
		select {
		// 刷新时间
		case <-r.refreshCtx.Done():
			continue
		case <-time.After(time.Minute * 5):
			_ = r.Shutdown()
			return
		}
	}
}

func (r *Room) Subscribe(c *wes.Connection) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.subs[c]; ok {
		return errors.New("already in room")
	}
	r.subs[c] = struct{}{}
	userId, _, _ := c.UserInfo()
	if userId == 0 {
		return errors.New("without user info")
	}
	ownerId, _, _ := r.owner.UserInfo()
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("user %d get in room of %d", userId, ownerId))
	c.DoneHook(func() {
		_ = r.UnSubscribe(c)
	})
	return nil
}

// UnSubscribe 退出房间
func (r *Room) UnSubscribe(c *wes.Connection) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.subs, c)
	ownerId, _, _ := r.owner.UserInfo()
	userId, _, _ := c.UserInfo()
	// 全部退出后关闭room
	if len(r.subs) == 0 {
		r.shutdownFree()
	}
	loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d exit room of %d", userId, ownerId))
	if c == r.owner && len(r.subs) != 0 {
		// 推举下一个房主
		for ele, _ := range r.subs {
			r.owner = ele
			break
		}
	}
	return nil
}

// Publish 向房间内所有成员发送消息
func (r *Room) Publish(v string, sender *wes.Connection) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.closed {
		return errors.New("room is closed")
	}
	if v == "" {
		return errors.New("msg is empty")
	}
	// 刷新room时间
	r.refresh()
	ctx, cancel := context.WithCancel(r.refreshCtx)
	r.refreshCtx = ctx
	r.refresh = cancel
	go r.closer()

	var shouldDelete = make([]*wes.Connection, 0)
	var senderName = ""
	var senderId int64 = 0
	if sender != nil {
		senderId, senderName, _ = sender.UserInfo()
		if senderId == 0 {
			return errors.New("sender is nil")
		}
	}
	var res = resp{
		SenderId:   senderId,
		SenderName: senderName,
		Timestamp:  time.Now().UnixMilli(),
		Data:       v,
	}
	data, _ := json.Marshal(res)
	for c := range r.subs {
		// 不向发送者发送消息
		if c == sender {
			continue
		}

		err := c.Send(data)
		if err != nil {
			shouldDelete = append(shouldDelete, c)
		}
	}
	// 延迟删除发送失败连接
	go func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		for _, conn := range shouldDelete {
			//loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("delete subscribe \"%s\" from %s", conn.RemoteAddr().String(), p.Name))
			delete(r.subs, conn)
		}
	}()
	return nil
}

// Start 启动
func (r *Room) Start() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.closed {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())

	r.refreshCtx = ctx
	r.refresh = cancel
	go r.closer()
	return nil
}

// 无锁关闭room
func (r *Room) shutdownFree() {
	r.refresh()
	clear(r.subs)
	id, _, _ := r.owner.UserInfo()
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room uuid %s closed, owner id %d", r.uuid, id))
	SetRoom(r.uuid, nil)
	r.owner = nil
	r.closed = true
}

// Shutdown 关闭room
func (r *Room) Shutdown() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.closed {
		return nil
	}
	r.shutdownFree()
	return nil
}

// NewRoom 创建room并开启，已存在则返回false
func NewRoom(owner *wes.Connection) (*Room, error) {

	id, username, _ := owner.UserInfo()
	if id == 0 {
		return nil, fmt.Errorf("no user info")
	}
	roomName := uuid.New().String()
	r := &Room{
		uuid:  roomName,
		subs:  make(map[*wes.Connection]struct{}),
		owner: owner,
		lock:  sync.RWMutex{},
	}
	ok := SetRoom(roomName, r)
	if !ok {
		return nil, fmt.Errorf("room exsit")
	}
	_ = r.Subscribe(owner)

	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room created by user %s id %d, room uuid %s", username, id, roomName))
	_ = r.Start()
	return r, nil
}

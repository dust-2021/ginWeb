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

type roomInfo struct {
	RoomID      string `json:"roomId"`
	RoomTitle   string `json:"roomTitle"`
	OwnerID     int64  `json:"ownerId"`
	OwnerName   string `json:"ownerName"`
	MemberCount int    `json:"memberCount"`
}

// RoomInfo 返回所有房间信息
func RoomInfo() []roomInfo {
	roomsLock.RLock()
	defer roomsLock.RUnlock()
	var infos []roomInfo
	for _, room := range rooms {

		item := roomInfo{
			RoomID:      room.uuid,
			RoomTitle:   room.Title,
			OwnerID:     room.Owner.UserId,
			OwnerName:   room.Owner.UserName,
			MemberCount: len(room.subs),
		}
		infos = append(infos, item)
	}
	return infos
}

type Room struct {
	uuid  string
	Title string
	subs  map[*wes.Connection]struct{}
	lock  sync.RWMutex
	Owner *wes.Connection

	refreshCtx context.Context
	refresh    context.CancelFunc
	closed     bool
}

func (r *Room) UUID() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.uuid
}

func (r *Room) ExistMember(c *wes.Connection) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.subs[c]
	return ok
}

type MateInfo struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Addr  string `json:"addr"`
	Owner bool   `json:"owner"`
}

// Mates 所有成员
func (r *Room) Mates() []MateInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	resp := make([]MateInfo, 0)
	for c := range r.subs {
		resp = append(resp, MateInfo{
			Name:  c.UserName,
			Id:    strconv.FormatInt(c.UserId, 10),
			Addr:  c.RemoteAddr().String(),
			Owner: c == r.Owner,
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

// Subscribe 订阅房间
func (r *Room) Subscribe(c *wes.Connection) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.subs[c]; ok {
		return nil
	}
	r.subs[c] = struct{}{}
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("user %d get in room of %d", c.UserId, r.Owner.UserId))
	c.DoneHook(func() {
		_ = r.UnSubscribe(c)
	})
	return nil
}

// UnSubscribe 退出房间
func (r *Room) UnSubscribe(c *wes.Connection) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.subs[c]; !ok {
		return nil
	}
	loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d exit room of %d", c.UserId, r.Owner.UserId))
	delete(r.subs, c)
	// 全部退出后关闭room
	if len(r.subs) == 0 {
		r.shutdownFree()
	}
	if c == r.Owner && len(r.subs) != 0 {
		// 推举下一个房主
		for ele, _ := range r.subs {
			r.Owner = ele
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
	var res = resp{
		SenderId:   sender.UserId,
		SenderName: sender.UserName,
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
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room uuid %s closed, owner id %d", r.uuid, r.Owner.UserId))
	SetRoom(r.uuid, nil)
	r.Owner = nil
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
func NewRoom(owner *wes.Connection, title string) (*Room, error) {
	roomName := uuid.New().String()
	r := &Room{
		uuid:  roomName,
		Title: title,
		subs:  make(map[*wes.Connection]struct{}),
		Owner: owner,
		lock:  sync.RWMutex{},
	}
	ok := SetRoom(roomName, r)
	if !ok {
		return nil, fmt.Errorf("room exsit")
	}
	_ = r.Subscribe(owner)

	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room created by user %s id %d, room uuid %s", owner.UserName, owner.UserId, roomName))
	_ = r.Start()
	return r, nil
}

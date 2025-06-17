package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"github.com/google/uuid"
	"sync"
	"time"
)

// Roomer 房间管理器单例
var Roomer *roomManager

// roomManager 房间管理类
type roomManager struct {
	rooms     map[string]*Room
	roomIndex []string // 排序器
	lock      sync.RWMutex
}

func (r *roomManager) Size() int {
	return len(Roomer.roomIndex)
}

func (r *roomManager) Get(name string) (*Room, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	v, ok := r.rooms[name]
	return v, ok
}

func (r *roomManager) removeIndex(key string) {
	var index []string
	for i, v := range r.roomIndex {
		if v == key {
			if i == len(r.roomIndex)-1 {
				r.roomIndex = index
			} else {
				r.roomIndex = append(index, r.roomIndex[i+1:]...)
			}
			return
		}
		index = append(index, v)
	}
	r.roomIndex = index
}

func (r *roomManager) Set(name string, room *Room) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	if room == nil {
		delete(r.rooms, name)
		r.removeIndex(name)
		return true
	}
	if _, ok := r.rooms[name]; ok {
		return false
	}
	r.rooms[name] = room
	r.roomIndex = append(r.roomIndex, name)
	return true
}

func (r *roomManager) Del(name string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.rooms, name)
}

func (r *roomManager) List(page int, size int) []RoomInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	infos := make([]RoomInfo, 0)
	if len(r.roomIndex) == 0 {
		return infos
	}
	start := (page - 1) * size
	end := start + size
	if end > len(r.roomIndex) {
		end = len(r.roomIndex)
	}
	for _, key := range r.roomIndex[start:end] {
		room, ok := r.rooms[key]
		if !ok {
			continue
		}
		item := RoomInfo{
			RoomID:       room.uuid,
			RoomTitle:    room.Config.Title,
			Description:  room.Config.Description,
			OwnerID:      room.Owner.UserId,
			OwnerName:    room.Owner.UserName,
			MemberCount:  len(room.subs),
			MaxMember:    room.Config.MaxMember,
			WithPassword: room.Config.Password != "",
			Forbidden:    room.forbidden,
		}
		infos = append(infos, item)
	}
	return infos
}

type RoomInfo struct {
	RoomID       string `json:"roomId"`
	RoomTitle    string `json:"roomTitle"`
	Description  string `json:"description"`
	OwnerID      int64  `json:"ownerId"`
	OwnerName    string `json:"ownerName"`
	MemberCount  int    `json:"memberCount"`
	MaxMember    int    `json:"memberMax"`
	WithPassword bool   `json:"withPassword"`
	Forbidden    bool   `json:"forbidden"`
}

// RoomConfig 房间设置
type RoomConfig struct {
	Title           string   `json:"title" validate:"required,max=12,min=2"` // 标题
	Description     string   `json:"description" validate:"max=128"`         // 描述
	MaxMember       int      `json:"maxMember" validate:"gte=1,lte=32"`      // 最大成员数
	Password        string   `json:"password" validate:"max=16,min=6"`       // 房间密码
	IPBlackList     []string `json:"blackList"`                              // ip黑名单
	UserIdBlackList []int64  `json:"UserIdBlackList"`                        // id黑名单
	DeviceBlackList []string `json:"deviceBlackList"`                        // 设备黑名单
	AutoClose       bool     `json:"autoClose"`                              // 是否自动关闭
}

type Room struct {
	uuid   string                       // id
	subs   map[*wes.Connection]struct{} // 成员ws连接对象
	lock   sync.RWMutex                 // 对象读写锁
	Owner  *wes.Connection              // 房间持有者
	Config *RoomConfig                  `json:"config"` //房间设置

	refreshCtx context.Context // 房间生命周期刷新上下文
	refresh    context.CancelFunc
	forbidden  bool // 房间是否关闭入口
	closed     bool // 对象状态
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
	Id    int    `json:"id"`
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
			Id:    int(c.UserId),
			Addr:  c.RemoteAddr().String(),
			Owner: c == r.Owner,
		})
	}
	return resp
}

// 生命周期管理
func (r *Room) closer() {
	// 未设置自动管理关闭直接退出
	if !r.Config.AutoClose {
		return
	}
	for {
		select {
		// ctx被主动取消则刷新时间重新计时，存在同时读写问题
		case <-r.refreshCtx.Done():
			continue
		// 计时结束，关闭房间
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
	if r.forbidden {
		return errors.New("room forbidden")
	}
	if r.Config.MaxMember != 0 && len(r.subs) >= r.Config.MaxMember {
		return errors.New("room is full")
	}
	for _, ip := range r.Config.IPBlackList {
		if ip == c.RemoteAddr().String() {
			return errors.New("black ip")
		}
	}
	for _, id := range r.Config.UserIdBlackList {
		if id == c.UserId {
			return errors.New("black user id")
		}
	}
	r.subs[c] = struct{}{}
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("user %d get in room of %d", c.UserId, r.Owner.UserId))
	// 将退出房间添加打ws连接关闭钩子中
	c.DoneHook(func() {
		_ = r.UnSubscribe(c)
	})
	r.Config.MaxMember += 1

	var res = wes.Resp{
		Id:         r.uuid,
		Method:     "publish.room.in",
		StatusCode: dataType.Success,
		Data: MateInfo{
			Id:    int(c.UserId),
			Name:  c.UserName,
			Owner: false,
			Addr:  c.RemoteAddr().String(),
		},
	}
	data, _ := json.Marshal(res)
	for conn := range r.subs {
		// 不向发送者发送消息
		if c == conn {
			continue
		}
		_ = conn.Send(data)
	}

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
		return nil
	}
	if c == r.Owner {
		// 推举下一个房主
		for ele, _ := range r.subs {
			r.Owner = ele
			break
		}
	}
	return nil
}

func (r *Room) Forbidden(to bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.forbidden = to
}

// Publish 向房间内所有成员发送消息
func (r *Room) Publish(v string, sender *wes.Connection) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if v == "" {
		return errors.New("msg is empty")
	}
	if r.Config.AutoClose {
		// 计时器
		r.refresh()
		ctx, cancel := context.WithCancel(r.refreshCtx)
		r.refreshCtx = ctx
		r.refresh = cancel
		go r.closer()
	}
	var res = wes.Resp{
		Id:         "",
		Method:     "publish.room.message",
		StatusCode: dataType.Success,
		Data: publisherResp{
			SenderId:   sender.UserId,
			SenderName: sender.UserName,
			Timestamp:  time.Now().UnixMilli(),
			Data:       v,
		},
	}
	data, _ := json.Marshal(res)
	for c := range r.subs {
		// 不向发送者发送消息
		if c == sender {
			continue
		}

		err := c.Send(data)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "WS ROOM", fmt.Sprintf("send to member %s failed", c.UserName))
		}
	}

	return nil
}

// Nat 成员一对一约定nat打洞
// to: 目标成员ip key: 唯一识别码
func (r *Room) Nat(to string, key string) {
	for c := range r.subs {
		if c.RemoteAddr().String() == to {
			resp := wes.Resp{
				Id:         "",
				Method:     "publish.room.nat",
				StatusCode: dataType.Success,
				Data:       key,
			}
			data, _ := json.Marshal(resp)
			err := c.Send(data)
			if err != nil {
				loguru.SimpleLog(loguru.Error, "WS ROOM", fmt.Sprintf("send nat msg to member %s failed", c.UserName))
			}
		}
	}
}

// Start 启动
func (r *Room) Start(timer string) error {
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

// 无锁关闭room，包内防止死锁
func (r *Room) shutdownFree() {
	r.refresh()
	clear(r.subs)
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room uuid %s closed, owner id %d", r.uuid, r.Owner.UserId))
	Roomer.Del(r.uuid)
	r.Owner = nil
	r.forbidden = true
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
func NewRoom(owner *wes.Connection, config *RoomConfig) (*Room, error) {
	roomName := uuid.New().String()
	r := &Room{
		uuid:   roomName,
		subs:   make(map[*wes.Connection]struct{}),
		Owner:  owner,
		lock:   sync.RWMutex{},
		Config: config,
	}
	ok := Roomer.Set(roomName, r)
	if !ok {
		return nil, fmt.Errorf("room exsit")
	}
	_ = r.Subscribe(owner)

	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room created by user %s id %d, room uuid %s", owner.UserName, owner.UserId, roomName))
	_ = r.Start("")
	return r, nil
}

func init() {
	Roomer = &roomManager{
		rooms: make(map[string]*Room),
	}
}

package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/loguru"
	"sync"
	"time"

	"github.com/google/uuid"
)

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// |                                          相关数据类型                                           |
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// MateInfo 接口返回的房间成员信息
type MateInfo struct {
	Name  string `json:"name"`
	Id    int    `json:"id"`
	Addr  string `json:"addr"`
	Owner bool   `json:"owner"`
	Vlan  string `json:"vlan"`
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

type mateAttr struct {
	Vlan uint16
}

// RoomConfig 房间设置
type RoomConfig struct {
	Title           string   `json:"title" validate:"required,max=12,min=2"`               // 标题
	Description     string   `json:"description" validate:"max=128"`                       // 描述
	MaxMember       int      `json:"maxMember" validate:"gte=1,lte=256"`                   // 最大成员数
	Password        *string  `json:"password,omitempty" validate:"omitempty,max=16,min=6"` // 房间密码
	IPBlackList     []string `json:"blackList"`                                            // ip黑名单
	UserIdBlackList []int64  `json:"UserIdBlackList"`                                      // id黑名单
	DeviceBlackList []string `json:"deviceBlackList"`                                      // 设备黑名单
	AutoClose       bool     `json:"autoClose"`                                            // 是否自动关闭
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// |                                           房间管理器                                            |
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Roomer 房间管理器单例
var Roomer = &roomManager{
	rooms: make(map[string]*room), roomIndex: make([]string, 0), lock: sync.RWMutex{},
}

// roomManager 房间管理类
type roomManager struct {
	rooms     map[string]*room
	roomIndex []string // 排序器
	lock      sync.RWMutex
}

func (r *roomManager) NewRoom(owner *wes.Connection, config *RoomConfig) (*room, error) {
	roomName := uuid.New().String()
	newRoom := &room{
		uuid:      roomName,
		subs:      make(map[*wes.Connection]mateAttr),
		Owner:     owner,
		lock:      sync.RWMutex{},
		Config:    config,
		vlan:      make(chan uint16, 1<<16),
		vlanYield: 0,
	}
	connVlan, err := newRoom.yieldVlan()
	if err != nil {
		return nil, err
	}
	newRoom.subs[owner] = mateAttr{Vlan: connVlan}
	owner.DoneHook("publish.room."+newRoom.uuid, func() {
		newRoom.lock.Lock()
		defer newRoom.lock.Unlock()
		loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d force to exit room of %d by done hook",
			owner.UserId, newRoom.Owner.UserId))
		// 最后执行删除，防止房间关闭导致空指针访问
		newRoom.deleteMember(owner)
	})
	_ = r.Set(roomName, newRoom)

	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room created by user %s id %d, room uuid %s", owner.UserName, owner.UserId, roomName))
	_ = newRoom.Start("")
	return newRoom, nil
}

func (r *roomManager) Size() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(Roomer.roomIndex)
}

func (r *roomManager) Get(name string) (*room, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	v, ok := r.rooms[name]
	return v, ok
}

func (r *roomManager) removeIndex(key string) {
	for i, v := range r.roomIndex {
		if v == key {
			if i == len(r.roomIndex)-1 {
				r.roomIndex = r.roomIndex[:i]
			} else {
				r.roomIndex = append(r.roomIndex[:i], r.roomIndex[i+1:]...)
			}
			return
		}
	}
}

// Set 添加房间，已存在同名房间则返回false
func (r *roomManager) Set(name string, room *room) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
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
	r.removeIndex(name)
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
	// 起点超限返回空
	if start >= len(r.roomIndex) {
		return infos
	}
	// 终点超限截断
	if end > len(r.roomIndex) {
		end = len(r.roomIndex)
	}
	for _, key := range r.roomIndex[start:end] {
		room_, ok := r.rooms[key]
		if !ok {
			continue
		}
		item := RoomInfo{
			RoomID:       room_.uuid,
			RoomTitle:    room_.Config.Title,
			Description:  room_.Config.Description,
			OwnerID:      room_.Owner.UserId,
			OwnerName:    room_.Owner.UserName,
			MemberCount:  len(room_.subs),
			MaxMember:    room_.Config.MaxMember,
			WithPassword: *room_.Config.Password != "",
			Forbidden:    room_.forbidden,
		}
		infos = append(infos, item)
	}
	return infos
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// |                                           房间对象                                              |
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type room struct {
	uuid   string                       // id
	subs   map[*wes.Connection]mateAttr // 成员ws连接对象
	lock   sync.RWMutex                 // 对象读写锁
	Owner  *wes.Connection              // 房间持有者
	Config *RoomConfig                  `json:"config"` //房间设置

	refreshCtx context.Context // 房间生命周期刷新上下文
	refresh    context.CancelFunc
	forbidden  bool // 房间是否关闭入口

	lifetimeCtx context.Context
	lifetimeEnd context.CancelFunc

	vlan      chan uint16 // 分配给每个成员的vlan IP后两段
	vlanYield int
}

func (r *room) UUID() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.uuid
}

func (r *room) ExistMember(c *wes.Connection) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.subs[c]
	return ok
}

// Mates 所有成员
func (r *room) Mates() []MateInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	resp := make([]MateInfo, 0)
	for c, attr := range r.subs {
		resp = append(resp, MateInfo{
			Name:  c.UserName,
			Id:    int(c.UserId),
			Addr:  c.IP,
			Owner: c == r.Owner,
			Vlan:  fmt.Sprintf("%d.%d", attr.Vlan>>8, attr.Vlan-(attr.Vlan>>8)),
		})
	}
	return resp
}

// 获取分配的IP段，已分配完时等待回收
func (r *room) yieldVlan() (vlan uint16, err error) {
	if r.vlanYield < 65536 {
		vlan = uint16(r.vlanYield)
		r.vlanYield++
		return
	}
	select {
	case vlan = <-r.vlan:
		return
	case <-time.After(time.Second * 5):
		err = errors.New("vlanYield timeout")
		return 0, err
	}
}

// 生命周期管理，维持一个计时器，自动关闭设置开启时生效，publish方法被调用后刷新计时器
func (r *room) closer() {
	// 未设置自动管理关闭直接退出
	if !r.Config.AutoClose {
		return
	}
	for {
		select {
		// ctx被主动取消则刷新时间重新计时
		case <-r.refreshCtx.Done():
			r.lock.Lock()
			ctx, cancel := context.WithCancel(context.Background())
			r.refreshCtx = ctx
			r.refresh = cancel
			r.lock.Unlock()
			continue
		// 计时结束，关闭房间
		case <-time.After(time.Minute * 30):
			_ = r.Shutdown()
			return
		case <-r.lifetimeCtx.Done():
			return
		}
	}
}

// Subscribe 订阅房间
func (r *room) Subscribe(c *wes.Connection) error {
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
		if ip == c.IP {
			return errors.New("black ip")
		}
	}
	for _, id := range r.Config.UserIdBlackList {
		if id == c.UserId {
			return errors.New("black user id")
		}
	}
	connVlan, err := r.yieldVlan()
	if err != nil {
		return err
	}
	r.subs[c] = mateAttr{Vlan: connVlan}
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("user %d get in room of %d", c.UserId, r.Owner.UserId))
	// 将退出房间添加打ws连接关闭钩子中
	c.DoneHook("publish.room."+r.uuid, func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d force to exit room of %d by done hook",
			c.UserId, r.Owner.UserId))
		// 最后执行删除，防止房间关闭导致空指针访问
		r.deleteMember(c)
	})
	go r.Notice(MateInfo{
		Id:    int(c.UserId),
		Name:  c.UserName,
		Owner: false,
		Addr:  c.IP,
	}, "in", c)

	return nil
}

// 删除成员并检测房间成员数量和房主转移
func (r *room) deleteMember(c *wes.Connection) {
	// 回收分配的IP网段
	r.vlan <- r.subs[c].Vlan
	delete(r.subs, c)
	// 全部退出后关闭room
	if len(r.subs) == 0 {
		r.shutdownFree()
	}
	go r.Notice(c.UserId, "out", c)
	if c == r.Owner {
		// 推举下一个房主
		for ele := range r.subs {
			r.Owner = ele
			go func() {
				type temp struct {
					Old int `json:"old"`
					New int `json:"new"`
				}
				r.Notice(temp{Old: int(c.UserId), New: int(r.Owner.UserId)}, "exchangeOwner", nil)
			}()
			break
		}
	}
}

// UnSubscribe 退出房间
func (r *room) UnSubscribe(c *wes.Connection) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.subs[c]; !ok {
		return nil
	}
	loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d exit room of %d", c.UserId, r.Owner.UserId))
	r.deleteMember(c)
	c.DeleteDoneHook("publish.room." + r.uuid)
	return nil
}

// Forbidden 房间禁止进入
func (r *room) Forbidden(to bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.forbidden = to
	go r.Notice(to, "forbidden", nil)
}

// Notice 发送系统通知，sender为通知触发者，不会收到消息，不会显式出现在报文中
func (r *room) Notice(v interface{}, type_ string, sender *wes.Connection) {
	note := "publish.room.notice"
	if len(type_) > 0 {
		note += "." + type_
	}
	var res = wes.Resp{
		Id:         r.uuid,
		Method:     note,
		StatusCode: dataType.Success,
		Data:       v,
	}
	data, _ := json.Marshal(res)
	err := r.Publish(data, sender)
	if err != nil {
		loguru.SimpleLog(loguru.Error, "WS ROOM", fmt.Sprintf("room notice of %s err: %v", type_, err))
	}
}

// Message 房间内发送消息
func (r *room) Message(msg string, sender *wes.Connection) {
	var res = wes.Resp{
		Id:         r.uuid,
		Method:     "publish.room.message",
		StatusCode: dataType.Success,
		Data: publisherResp{
			SenderId:   sender.UserId,
			SenderName: sender.UserName,
			Timestamp:  time.Now().UnixMilli(),
			Data:       msg,
		},
	}
	data, _ := json.Marshal(res)
	go func() {
		_ = r.Publish(data, sender)
	}()

}

// Publish 向所有成员广播消息，提供sender后不向sender发送
func (r *room) Publish(v []byte, sender *wes.Connection) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if len(v) == 0 {
		return errors.New("msg is empty")
	}
	if r.Config.AutoClose {
		// 计时器
		r.refresh()
	}
	for c := range r.subs {
		// 不向发送者发送消息
		if c == sender {
			continue
		}

		err := c.Send(v)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "WS ROOM", fmt.Sprintf("send to member %s failed", c.UserName))
		}
	}

	return nil
}

// Nat 成员一对一约定nat打洞
// to: 目标成员ip，为“*”时指向所有成员 key: 唯一识别码
func (r *room) Nat(to string, key string) {
	for c := range r.subs {
		if to == "*" || c.IP == to {
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
func (r *room) Start(...interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.lifetimeCtx, r.lifetimeEnd = context.WithCancel(context.Background())
	if r.Config.AutoClose {
		ctx, cancel := context.WithCancel(context.Background())

		r.refreshCtx = ctx
		r.refresh = cancel
		go r.closer()
	}
	return nil
}

// 无锁关闭room，包内防止死锁
func (r *room) shutdownFree() {
	clear(r.subs)
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room uuid %s closed, owner id %d", r.uuid, r.Owner.UserId))
	Roomer.Del(r.uuid)
	r.lifetimeEnd()
}

// Shutdown 关闭room
func (r *room) Shutdown() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Notice("", "close", nil)
	r.shutdownFree()
	return nil
}

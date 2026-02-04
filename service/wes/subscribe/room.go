package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wireguard"
	"ginWeb/utils/loguru"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// |                                          相关数据类型                                           |
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// MateInfo 接口返回的房间成员信息
type MateInfo struct {
	Name      string `json:"name"`
	Uuid      string `json:"uuid"`
	Id        int    `json:"id"`
	Owner     bool   `json:"owner"`
	Vlan      int    `json:"vlan"`
	PublicKey string `json:"publicKey"`
	WgIp      string `json:"wgIp"`    // 成员真实IP
	WgPort    int    `json:"wgPort"`  // 成员真实端口
	UdpPort   int    `json:"udpPort"` // 成员本地udp端口
}

// 用于接收创建房间数据
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

// notice通知返回的成员地址变动信息
type NoticeUpdateEndpoint struct {
	Uuid string `json:"uuid"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type mateAttr struct {
	Vlan      uint16 // 分配的虚拟局域网网段号
	PublicKey string // ed25519生成的32位公钥，用于vlan通信
	WgIP      string // 成员真实wg外网ip
	WgPort    int    // 成员真实wg外网端口
	UdpPort   int    // 成员本地udp端口
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

// 创建时传入owner信息
func (r *roomManager) NewRoom(owner *wes.Connection, config *RoomConfig, args ...any) (*room, error) {
	roomName := uuid.NewString()
	newRoom := &room{
		uuid:      roomName,
		Link:      uuid.NewString(),
		subs:      make(map[*wes.Connection]mateAttr),
		ownerConn: owner,
		lock:      sync.RWMutex{},
		Config:    config,
		forbidden: true,
	}
	connVlan, err := wireguard.WireguardManager.AddPeer(owner.Uuid, args[0].(string), newRoom.UpdateTrueAddr)
	if err != nil {
		return nil, err
	}
	newRoom.subs[owner] = mateAttr{Vlan: connVlan, PublicKey: args[0].(string), UdpPort: args[1].(int)}
	// 将退出房间添加到ws连接关闭钩子中，主动退出房间将会删除该钩子
	owner.DoneHook("publish.room."+newRoom.uuid, func() {
		wireguard.WireguardManager.RemovePeer(owner.Uuid)
		newRoom.lock.Lock()
		defer newRoom.lock.Unlock()
		loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d force to exit room %s by done hook",
			owner.UserId, newRoom.uuid))
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

func (r *roomManager) Get(roomUidOrLink string) (*room, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	v, ok := r.rooms[roomUidOrLink]
	if ok {
		return v, ok
	}
	for _, room := range r.rooms {
		if room.Link == roomUidOrLink {
			return room, true
		}
	}
	return nil, false
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
func (r *roomManager) Set(roomUid string, room *room) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.rooms[roomUid]; ok {
		return false
	}
	r.rooms[roomUid] = room
	r.roomIndex = append(r.roomIndex, roomUid)
	return true
}

func (r *roomManager) Del(roomUid string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.rooms, roomUid)
	r.removeIndex(roomUid)
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
		infos = append(infos, room_.Info())
	}
	return infos
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// |                                           房间对象                                              |
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type room struct {
	uuid      string                       // id
	subs      map[*wes.Connection]mateAttr // 成员ws连接对象
	lock      sync.RWMutex                 // 对象读写锁
	ownerConn *wes.Connection              // 房间持有者
	Link      string                       // 无视关闭状态和密码的进房链接
	Config    *RoomConfig                  `json:"config"` //房间设置

	refreshCtx context.Context // 房间生命周期刷新上下文
	refresh    context.CancelFunc
	forbidden  bool // 房间是否关闭入口

	lifetimeCtx context.Context
	lifetimeEnd context.CancelFunc
}

func (r *room) Info() RoomInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return RoomInfo{
		RoomID:       r.uuid,
		RoomTitle:    r.Config.Title,
		Description:  r.Config.Description,
		OwnerID:      r.ownerConn.UserId,
		OwnerName:    r.ownerConn.UserName,
		MemberCount:  len(r.subs),
		MaxMember:    r.Config.MaxMember,
		WithPassword: r.Config.Password != nil && *r.Config.Password != "",
		Forbidden:    r.forbidden,
	}
}

func (r *room) UUID() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.uuid
}

func (r *room) OwnerUuid() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.ownerConn.Uuid
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
			Name:      c.UserName,
			Uuid:      c.UserUuid,
			Id:        int(c.UserId),
			Owner:     c == r.ownerConn,
			Vlan:      int(attr.Vlan),
			PublicKey: attr.PublicKey,
			WgIp:      attr.WgIP,
			WgPort:    attr.WgPort,
			UdpPort:   attr.UdpPort,
		})
	}
	return resp
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

// 用于wg更新peer真实地址的钩子函数
func (r *room) UpdateTrueAddr(uid string, ip string, port int) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for c := range r.subs {
		if c.Uuid != uid {
			continue
		}
		if r.subs[c].WgIP == ip && r.subs[c].WgPort == port {
			return
		}
		info := r.subs[c]
		info.WgIP = ip
		info.WgPort = port
		r.subs[c] = info
		loguru.SimpleLog(loguru.Debug, "ROOM", fmt.Sprintf("peer wg address update to %s:%d", ip, port))
		go r.Notice(NoticeUpdateEndpoint{Uuid: c.UserUuid, Ip: ip, Port: port}, "updatePeerEndpoint", c)
		return
	}

}

// Subscribe 订阅房间，订阅时传入publicKey和本地udp端口
func (r *room) Subscribe(c *wes.Connection, args ...any) error {
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
	if slices.Contains(r.Config.IPBlackList, c.IP) {
		return errors.New("black ip")
	}
	for _, id := range r.Config.UserIdBlackList {
		if id == c.UserId {
			return errors.New("black user id")
		}
	}
	connVlan, err := wireguard.WireguardManager.AddPeer(c.Uuid, args[0].(string), r.UpdateTrueAddr)
	if err != nil {
		return err
	}
	r.subs[c] = mateAttr{Vlan: connVlan, UdpPort: args[1].(int), PublicKey: args[0].(string)}
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("user %d get in room %s", c.UserId, r.uuid))
	// 将退出房间添加到ws连接关闭钩子中，主动退出房间将会删除该钩子
	c.DoneHook("publish.room."+r.uuid, func() {
		wireguard.WireguardManager.RemovePeer(c.Uuid)
		r.lock.Lock()
		defer r.lock.Unlock()
		loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %d force to exit room %s by done hook",
			c.UserId, r.uuid))
		// 最后执行删除，防止房间关闭导致空指针访问
		r.deleteMember(c)
	})
	go r.Notice(MateInfo{
		Id:        int(c.UserId),
		Name:      c.UserName,
		Uuid:      c.UserUuid,
		Owner:     false,
		Vlan:      int(connVlan),
		PublicKey: args[0].(string),
		UdpPort:   args[1].(int),
	}, "in", c)

	return nil
}

// 删除成员并检测房间成员数量和房主转移
func (r *room) deleteMember(c *wes.Connection) {
	delete(r.subs, c)
	// 全部退出后关闭room
	if len(r.subs) == 0 {
		r.shutdownFree()
	}
	go r.Notice(c.UserUuid, "out", c)
	if c == r.ownerConn {
		// 推举下一个房主
		for ele := range r.subs {
			r.ownerConn = ele
			go func() {
				type temp struct {
					Old string `json:"old"`
					New string `json:"new"`
				}
				r.Notice(temp{Old: c.UserUuid, New: r.ownerConn.UserUuid}, "exchangeOwner", nil)
			}()
			break
		}
	}
}

// UnSubscribe 退出房间
func (r *room) UnSubscribe(c *wes.Connection, args ...any) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.subs[c]; !ok {
		return nil
	}
	loguru.SimpleLog(loguru.Debug, "WS ROOM", fmt.Sprintf("user %s exit room of %s", c.UserUuid, r.uuid))
	r.deleteMember(c)
	// wireguard中删除peer
	wireguard.WireguardManager.RemovePeer(c.Uuid)
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
	if type_ != "" {
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
			SenderUuid: sender.Uuid,
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
			loguru.SimpleLog(loguru.Error, "WS ROOM", fmt.Sprintf("send to member %s failed", c.UserUuid))
		}
	}

	return nil
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
	loguru.SimpleLog(loguru.Info, "WS ROOM", fmt.Sprintf("room uuid %s closed", r.uuid))
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

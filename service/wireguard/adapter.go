package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/curve25519"

	"ginWeb/config"
	"ginWeb/utils/loguru"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
)

// 负责管理wireguard配置
var WireguardManager *adapterManager = nil

type adapter struct {
	publicKey  []byte
	privateKey []byte
	listenPort uint16
}

type peer struct {
	RoomID    string
	PublicKey [device.NoisePublicKeySize]byte
	Vlan      uint16
	WgPeer    *device.Peer
}

// 适配器管理
type adapterManager struct {
	lock        *sync.RWMutex
	wgInterface adapter
	peers       map[string]map[string]*peer
	vlanID      uint16
	vlanRecover chan uint16
}

func (am *adapterManager) PeersCount() int {
	return len(am.peers)
}

func (am *adapterManager) config() string {
	return fmt.Sprintf("private_key=%x\r\nlisten_port=%d\r\nreplace_peers=true",
		am.wgInterface.privateKey, am.wgInterface.listenPort)
}

// base64编码公钥
func (am *adapterManager) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(am.wgInterface.publicKey)
}

func (am *adapterManager) Start() (err error) {
	err = server.Open()
	if err != nil {
		return fmt.Errorf("open wireguard tun device failed-%s", err.Error())
	}
	err = server.Device.IpcSet(am.config())
	if err != nil {
		return fmt.Errorf("set wireguard interface failed-%s", err.Error())
	}
	return server.Device.Up()
}

// 分配局域网IPV4后两段，最多支持65535-1个局域网IP
func (am *adapterManager) vlan() (v uint16, err error) {
	if len(am.vlanRecover) != 0 {
		return <-am.vlanRecover, nil
	}
	if am.vlanID < 65535 {
		am.vlanID++
		return am.vlanID, nil
	}
	select {
	case v := <-am.vlanRecover:
		return v, nil
	case <-time.After(10 * time.Second):
		return 0, fmt.Errorf("no available vlan id")
	}

}

func (am *adapterManager) Close() error {
	server.Device.Down()
	server.Close()
	return nil
}

// 添加局域网成员
func (am *adapterManager) AddPeer(roomID string, uname string, publicKey string) (vlan uint16, err error) {

	pubKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(pubKey) != device.NoisePublicKeySize {
		return 0, err
	}
	var pubByte device.NoisePublicKey
	copy(pubByte[:], pubKey)
	vlan_ip, err := am.vlan()
	if err != nil {
		return 0, err
	}
	vlan_ip_string := fmt.Sprintf("%d.%d.%d.%d/32", config.Conf.Server.Vlan[0], config.Conf.Server.Vlan[1], vlan_ip>>8, vlan_ip&0xff)
	wgPeer, err := server.Device.NewPeerFix(pubByte, vlan_ip_string, func(ip conn.Endpoint) {
		loguru.SimpleLog(loguru.Debug, "WG", fmt.Sprintf("peer %s is now at %s", uname, ip.DstToString()))
	})
	if err != nil {
		// 回收vlan地址
		am.vlanRecover <- vlan_ip
		return 0, err
	}
	curPeer := &peer{
		RoomID:    roomID,
		PublicKey: pubByte,
		Vlan:      vlan_ip,
		WgPeer:    wgPeer,
	}

	am.lock.Lock()
	defer am.lock.Unlock()
	if _, ok := am.peers[roomID]; !ok {
		am.peers[roomID] = make(map[string]*peer)
	}

	am.peers[roomID][uname] = curPeer
	loguru.SimpleLog(loguru.Info, "WG", fmt.Sprintf("add peer %s to room %s with vlan %s", uname, roomID, vlan_ip_string))
	return vlan_ip, nil
}

// 删除对等体并回收分发的网段IP
func (am *adapterManager) RemovePeer(roomId string, uname string) {
	am.lock.Lock()
	defer am.lock.Unlock()
	if _, ok := am.peers[roomId]; !ok {
		return
	}
	if _, ok := am.peers[roomId][uname]; !ok {
		return
	}
	server.Device.RemovePeer(am.peers[roomId][uname].PublicKey)
	am.vlanRecover <- am.peers[roomId][uname].Vlan
	loguru.SimpleLog(loguru.Info, "WG", fmt.Sprintf("remove peer %s from room %s with vlan %d", uname, roomId, am.peers[roomId][uname].Vlan))
	delete(am.peers[roomId], uname)
	if length := len(am.peers[roomId]); length == 0 {
		delete(am.peers, roomId)
	}
}

func (am *adapterManager) GetIpcConfig() (string, error) {
	return server.Device.IpcGet()
}

func init() {
	// 生成随机私钥
	var privateKey [32]byte
	_, err := rand.Read(privateKey[:])
	if err != nil {
		panic(fmt.Errorf("generate wg key error"))
	}
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	WireguardManager = &adapterManager{
		lock:  &sync.RWMutex{},
		peers: make(map[string]map[string]*peer),
		wgInterface: adapter{
			publicKey:  privateKey[:],
			privateKey: publicKey,
			listenPort: config.Conf.Server.UdpPort,
		},
		vlanID:      1,
		vlanRecover: make(chan uint16, 1<<16),
	}
	loguru.SimpleLog(loguru.Debug, "WG", fmt.Sprintf("generate wg pub key %s", WireguardManager.GetPublicKey()))
}

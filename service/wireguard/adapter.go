package wireguard

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"ginWeb/config"
	"ginWeb/utils/loguru"

	"golang.zx2c4.com/wireguard/device"
)

var WireguardManager *AdapterManager = nil

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
type AdapterManager struct {
	lock        *sync.RWMutex
	wgInterface adapter
	peers       map[string]map[string]*peer
	vlanID      uint16
	vlanRecover chan uint16
}

func (am *AdapterManager) PeersCount() int {
	return len(am.peers)
}

func (am *AdapterManager) config() string {
	return fmt.Sprintf("private_key=%x\r\nlisten_port=%d\r\nreplace_peers=true",
		am.wgInterface.privateKey, am.wgInterface.listenPort)
}

// base64编码公钥
func (am *AdapterManager) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(am.wgInterface.publicKey)
}

func (am *AdapterManager) Start() (err error) {
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

func (am *AdapterManager) vlan() (v uint16, err error) {
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

func (am *AdapterManager) Close() error {
	server.Device.Down()
	server.Close()
	return nil
}

// 添加局域网成员
func (am *AdapterManager) AddPeer(roomID string, uname string, publicKey string) (vlan uint16, err error) {

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
	vlan_ip_string := fmt.Sprintf("10.0.%d.%d/16", vlan_ip>>8, vlan_ip&0xff)
	wgPeer, err := server.Device.NewPeerFix(pubByte, vlan_ip_string, nil)
	if err != nil {
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
	if err != nil {
		am.vlanRecover <- vlan_ip
		return 0, err
	}
	loguru.SimpleLog(loguru.Info, "WG", fmt.Sprintf("add peer %s to room %s with vlan %d", uname, roomID, vlan_ip))
	return vlan_ip, nil
}

// 删除对等体并回收分发的网段IP
func (am *AdapterManager) RemovePeer(roomId string, uname string) {
	am.lock.Lock()
	defer am.lock.Unlock()
	if _, ok := am.peers[roomId]; !ok {
		return
	}
	if _, ok := am.peers[roomId][uname]; !ok {
		return
	}
	delete(am.peers[roomId], uname)
	if length := len(am.peers[roomId]); length == 0 {
		delete(am.peers, roomId)
	}
	server.Device.RemovePeer(am.peers[roomId][uname].PublicKey)
	am.vlanRecover <- am.peers[roomId][uname].Vlan
	loguru.SimpleLog(loguru.Info, "WG", fmt.Sprintf("remove peer %s from room %s with vlan %d", uname, roomId, am.peers[roomId][uname].Vlan))
}

func init() {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	WireguardManager = &AdapterManager{
		lock:  &sync.RWMutex{},
		peers: make(map[string]map[string]*peer),
		wgInterface: adapter{
			publicKey:  pub,
			privateKey: pri.Seed(),
			listenPort: config.Conf.Server.UdpPort,
		},
		vlanID:      1,
		vlanRecover: make(chan uint16, 1<<16),
	}
}

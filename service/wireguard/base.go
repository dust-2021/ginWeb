package wireguard

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"

	"ginWeb/config"
	"ginWeb/utils/loguru"
)

var server = &tunDevice{}

type tunDevice struct {
	Device *device.Device
	Tun    tun.Device
	Link   netlink.Link
}

func wgLogger(pre string, level string) func(format string, args ...any) {
	return func(format string, args ...any) {
		switch level {
		case "debug":
			loguru.Logger.Debugf(fmt.Sprintf("[WG-%s] | %s", pre, format), args...)
		case "error":
			loguru.Logger.Errorf(fmt.Sprintf("[WG-%s] | %s", pre, format), args...)
		default:
			loguru.Logger.Infof(fmt.Sprintf("[WG-%s] | %s", pre, format), args...)
		}
	}
}

func (ws *tunDevice) Open() (err error) {
	ws.Tun, err = tun.CreateTUN("wg0", device.DefaultMTU)
	if err != nil {
		return fmt.Errorf("create tun device failed: %v", err)
	}
	// 获取实际的设备名
	realName, err := ws.Tun.Name()
	if err != nil {
		ws.Tun.Close()
		return fmt.Errorf("init tun device failed: %v", err)
	}

	// 2. 创建 WireGuard 设备
	logger := &device.Logger{
		Verbosef: wgLogger(realName, "debug"),
		Errorf:   wgLogger(realName, "error"),
	}

	ws.Device = device.NewDevice(ws.Tun, conn.NewDefaultBind(), logger)

	ws.Link, err = netlink.LinkByName(realName)
	if err != nil {
		return
	}
	addr, err := netlink.ParseAddr(fmt.Sprintf("%d.%d.0.1/16", config.Conf.Server.Vlan[0], config.Conf.Server.Vlan[1]))
	if err != nil {
		return fmt.Errorf("parse wireguard vlan ip failed: %v", err)
	}
	if netlink.AddrAdd(ws.Link, addr) != nil {
		return fmt.Errorf("add wireguard vlan ip failed: %v", err)
	}
	return netlink.LinkSetUp(ws.Link)
}

func (ws *tunDevice) Close() {
	if ws.Device != nil {
		ws.Device.Close()
	}
	if ws.Tun != nil {
		ws.Tun.Close()
	}
}

package wireguard

import (
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"

	"ginWeb/utils/loguru"
)

var server = &tunDevice{}

type tunDevice struct {
	Device *device.Device
	Tun    tun.Device
}

func wgLogger(pre string, level string) func(format string, args ...any) {
	return func(format string, args ...any) {
		switch level {
		case "debug":
			loguru.Logger.Debugf(fmt.Sprintf("[WG-%s] | %s", pre, format), args...)
		case "error":
			loguru.Logger.Errorf(fmt.Sprintf("[WG-%s] | %s", pre, format), args...)
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
	return nil
}

func (ws *tunDevice) Close() {
	if ws.Device != nil {
		ws.Device.Close()
	}
	if ws.Tun != nil {
		ws.Tun.Close()
	}
}

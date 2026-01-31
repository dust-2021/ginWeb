package wireguard

import (
	"fmt"
	"os/exec"
	"strings"

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
	err = netlink.LinkSetUp(ws.Link)
	if err != nil {
		return fmt.Errorf("set wireguard link up failed: %v", err)
	}
	// err = enableIPForwarding()
	// if err != nil {
	// 	return fmt.Errorf("enable ip forwarding failed: %v", err)
	// }
	// err = setupIPTables(realName)
	// if err != nil {
	// 	return fmt.Errorf("setup iptables failed: %v", err)
	// }
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

// 启用内核 IP 转发
func enableIPForwarding() error {
	// 方法 1：直接写入 sysctl（立即生效）
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return err
	}

	// 方法 2：或者直接写文件
	// return os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)

	return nil
}

// 配置 iptables 转发规则
func setupIPTables(interfaceName string) error {
	rules := [][]string{
		// 允许 WireGuard 接口的转发流量
		{"-A", "FORWARD", "-i", interfaceName, "-j", "ACCEPT"},
		{"-A", "FORWARD", "-o", interfaceName, "-j", "ACCEPT"},

		// 允许已建立的连接
		// {"-A", "FORWARD", "-i", interfaceName, "-o", interfaceName, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"},
	}

	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		if output, err := cmd.CombinedOutput(); err != nil {
			// 忽略 "already exists" 错误
			if !strings.Contains(string(output), "already exists") {
				return fmt.Errorf("iptables rule failed: %v, output: %s", err, output)
			}
		}
	}

	return nil
}

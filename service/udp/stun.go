package udp

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/utils/loguru"
	"net"
	"strings"
)

// UdpSvr udp服务
var UdpSvr *udpService

type udpService struct {
	Port uint16
	Conn *net.UDPConn
}

func (s *udpService) handle(addr *net.UDPAddr, data string) {
	loguru.SimpleLog(loguru.Info, "NAT", fmt.Sprintf("receive from %s, data:%s", addr.String(), data))
	// 约定数据格式：[type:string]\r\n[uuid:string]\r\n([data:string])?
	resp := strings.Split(data, "\r\n")
	if len(resp) != 3 {
		return
	}
	switch resp[0] {
	// 接收客户端主副端口请求，并分别返回其对应的公网地址
	case "turn":
		_, err := s.Conn.WriteTo(fmt.Appendf(nil, "turn\r\n%s\r\n%s", resp[1], addr.String()), addr)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "NAT", err.Error())
		}
	default:
		return
	}
}

func (s *udpService) start() {
	buf := make([]byte, 1024)
	for {
		n, addr, err := s.Conn.ReadFromUDP(buf)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "NAT", err.Error())
		}
		if n == 1024 {
			loguru.SimpleLog(loguru.Warn, "NAT", "udp data too long")
			continue
		}
		go s.handle(addr, string(buf[:n]))
	}
}

func (s *udpService) Close() error {
	return s.Conn.Close()
}

func (s *udpService) Run() (err error) {
	s.Conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: int(s.Port),
	})
	go s.start()
	return
}

func init() {
	UdpSvr = &udpService{
		Port: config.Conf.Server.TurnPort,
	}
}

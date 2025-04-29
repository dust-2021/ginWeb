package udp

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/utils/loguru"
	"net"
	"strings"
	"sync"
	"time"
)

type signal struct {
	IP   string
	Port int
}

// Handler nat打洞的udp发包控制器
var Handler *natServer

type natServer struct {
	Port  uint16
	Conn  *net.UDPConn
	Stuns map[string]signal
	lock  sync.RWMutex
}

func (s *natServer) handle(addr *net.UDPAddr, data string) {
	loguru.SimpleLog(loguru.Info, "NAT", fmt.Sprintf("receive from %s, data:%s", addr.String(), data))
	resp := strings.Split(data, "\r\n")
	switch resp[0] {
	case "stun", "stunSlave":
		s.lock.RLock()
		defer s.lock.RUnlock()
		// resp[1]是该次stun测试的唯一Id
		v, ok := s.Stuns[resp[1]]
		if !ok {
			s.Stuns[resp[1]] = signal{addr.IP.String(), addr.Port}
			// 五秒后删除这个stun测试，不管是否完成
			time.AfterFunc(time.Second*5, func() {
				s.lock.Lock()
				defer s.lock.Unlock()
				delete(s.Stuns, resp[1])
			})
			break
		}
		data := fmt.Sprintf("stun\r\n%v", v.Port == addr.Port)
		_, err := s.Conn.WriteTo([]byte(data), &net.UDPAddr{IP: addr.IP, Port: v.Port})
		if err != nil {
			loguru.SimpleLog(loguru.Error, "NAT", err.Error())
		}
		delete(s.Stuns, resp[1])
		break
	case "heartbeat":
		// 心跳检测不处理，能接收到视为成功
		break
	case "ping":
		_, err := s.Conn.WriteTo([]byte("pong"), addr)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "NAT", err.Error())
		}
		break
	default:
		return
	}
}

func (s *natServer) start() {
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

func (s *natServer) Close() error {
	return s.Conn.Close()
}

func (s *natServer) Run() (err error) {
	s.Conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: int(s.Port),
	})
	go s.start()
	return
}

func init() {
	Handler = &natServer{
		Port:  config.Conf.Server.UdpPort,
		Stuns: map[string]signal{},
	}
}

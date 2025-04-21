package udp

import (
	"ginWeb/config"
	"ginWeb/utils/loguru"
	"net"
)

// Handler nat打洞的udp发包控制器
var Handler *udpServer

type udpServer struct {
	Port uint16
	Conn *net.UDPConn
}

func (s *udpServer) handle(data string) {

}

func (s *udpServer) start() {
	buf := make([]byte, 1024)
	for {
		n, _, err := s.Conn.ReadFromUDP(buf)
		if err != nil {
			loguru.SimpleLog(loguru.Error, "UDP", err.Error())
		}
		if n == 1024 {
			loguru.SimpleLog(loguru.Warn, "UDP", "udp data too long")
			continue
		}
		go s.handle(string(buf[:n]))
	}
}

func (s *udpServer) Close() error {
	return s.Conn.Close()
}

func (s *udpServer) Run() (err error) {
	s.Conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: int(s.Port),
	})
	go s.start()
	return
}

func init() {
	Handler = &udpServer{
		Port: config.Conf.Server.UdpPort,
	}
}

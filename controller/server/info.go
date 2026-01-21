package server

import (
	"ginWeb/config"
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"ginWeb/service/wireguard"
	"time"

	"github.com/gin-gonic/gin"
)

type wgInfo struct {
	PublicKey  string `json:"publicKey"`
	ListenPort uint16 `json:"listenPort"`
	VlanIp     [2]int `json:"vlanIp"`
}

type connInfo struct {
	ServerTime  int64 `json:"serverTime"`
	WsConnected int   `json:"wsConnected"`
	WgPeers     int   `json:"wgPeers"`
	Rooms       int   `json:"rooms"`
}

type InfoMessage struct{}

func (i InfoMessage) Connecting(ctx *gin.Context) {
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: connInfo{
			ServerTime:  time.Now().UnixMilli(),
			WsConnected: wes.ConnManager.Count(),
			WgPeers:     wireguard.WireguardManager.PeersCount(),
			Rooms:       subscribe.Roomer.Size(),
		},
	})
}

func (i InfoMessage) Wginfo(ctx *gin.Context) {
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: wgInfo{
			PublicKey:  wireguard.WireguardManager.GetPublicKey(),
			ListenPort: config.Conf.Server.UdpPort,
			VlanIp:     config.Conf.Server.Vlan,
		},
	})
}

func (i InfoMessage) RegisterRoute(r string, g *gin.RouterGroup) {
	group := g.Group(r)
	group.Handle("GET", "connecting", middleware.NewIndependentLimiter(1000, 0, 0).HttpHandle, i.Connecting)
	group.Handle("GET", "wginfo", middleware.NewIndependentLimiter(1000, 0, 0).HttpHandle, i.Wginfo)
}

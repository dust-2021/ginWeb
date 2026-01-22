package debug

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wireguard"

	"github.com/gin-gonic/gin"
)

type WgDebug struct {
}

func (w WgDebug) IpcConfig(ctx *gin.Context) {
	conf, err := wireguard.WireguardManager.GetIpcConfig()
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.Unknown,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: conf,
	})
}

func (w WgDebug) RegisterRoute(r string, g *gin.RouterGroup) {
	group := g.Group(r)
	group.Handle("GET", "/ipcConfig", w.IpcConfig)
}

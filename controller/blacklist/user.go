package blacklist

import (
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"

	"github.com/gin-gonic/gin"
)

type UserList struct{}

func (u UserList) Add(ctx *gin.Context) {
	target := ctx.Query("uuid")
	tokenS, f := ctx.Get("token")
	token, flag := tokenS.(*auth.Token)
	if !f || !flag {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.NoToken,
			Message: "no token",
		})
		return
	}
	if target == token.UserUUID {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: "can't add yourself to blacklist",
		})
		return
	}
	data := systemMode.UserBlacklist{
		UserUuid:   token.UserUUID,
		TargetUuid: target,
	}
	err := data.Add()
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.WrongData, Message: err.Error(),
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}

func (u UserList) Delete(ctx *gin.Context) {
	target := ctx.Query("uuid")
	tokenS, f := ctx.Get("token")
	token, flag := tokenS.(*auth.Token)
	if !f || !flag {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.NoToken,
			Message: "no token",
		})
		return
	}
	data := systemMode.UserBlacklist{
		UserUuid:   token.UserUUID,
		TargetUuid: target,
	}
	err := data.Delete()
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.WrongData, Message: err.Error(),
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}

func (u UserList) RegisterRoute(route string, group *gin.RouterGroup) {
	g := group.Group(route)
	g.Handle("GET", "add", u.Add)
	g.Handle("GET", "delete", u.Delete)
}

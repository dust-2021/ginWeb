package middleware

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"ginWeb/utils/tools"
	"github.com/gin-gonic/gin"
)

type permission struct {
	// 必要权限
	Permission []string
	// 可选权限，只用包含每个子数组中的一个
	SelectPermission [][]string
}

func (p *permission) handle(perm []string) bool {
	return tools.ContainsAll(perm, p.Permission) && tools.ContainOne(perm, p.SelectPermission)
}

func (p *permission) HttpHandle(c *gin.Context) {
	tokenPtr, f := c.Get("token")
	if !f {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: dataType.NoToken, Message: "without token",
		})
		return
	}
	token, flag := tokenPtr.(*auth.Token)
	if !flag {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: dataType.WrongToken, Message: "without token",
		})
		return
	}
	if !p.handle(token.Permission) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: dataType.DeniedByPermission, Message: "denied",
		})
	}

}

func (p *permission) WsHandle(c *wes.WContext) {
	if !p.handle(c.Conn.UserPermission) {
		c.Result(dataType.DeniedByPermission, "denied")
	}
}

// NewPermission 对比token中存储的权限是否足够，需要前置loginStatus中间件
func NewPermission(perms []string, choice ...[]string) Middleware {
	per := &permission{
		Permission:       perms,
		SelectPermission: choice,
	}
	return per
}

package ginMiddle

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
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

func (p *permission) Handle(c *gin.Context) {
	if len(p.Permission) == 0 {
		c.Next()
		return
	}
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
	if !(tools.ContainsAll(token.Permission, p.Permission) && tools.ContainOne(token.Permission, p.SelectPermission)) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: dataType.DeniedByPermission, Message: "permission don't match",
		})
	}

}

// NewPermission 对比token中存储的权限是否足够，需要前置loginStatus中间件
func NewPermission(perms []string, choice ...[]string) middleware.Middleware {
	per := &permission{
		Permission:       perms,
		SelectPermission: choice,
	}
	return per
}

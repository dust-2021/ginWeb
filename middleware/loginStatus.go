package middleware

import (
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

// LoginStatus token验证中间件，验证token是否正确、是否过期、是否已注销，并将token指针放入请求上下文中
type LoginStatus struct {
	Redirect bool
}

func (s *LoginStatus) Handle(c *gin.Context) {
	tokenStr := c.GetHeader("Token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dataType.JsonWrong{
			Code: 1, Message: "invalid token",
		})
		return
	}
	// 验证是否为黑名单Token
	_, err := reCache.Get("blackToken", tokenStr)
	if err == nil {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "invalid token",
		})
		return
	}

	token, err := auth.CheckToken(tokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	c.Set("token", token)
}

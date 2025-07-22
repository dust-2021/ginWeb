package middleware

import (
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type loginStatus struct {
}

func (l *loginStatus) HttpHandle(c *gin.Context) {
	tokenStr := c.GetHeader("Token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dataType.JsonWrong{
			Code: dataType.NoToken, Message: "invalid token",
		})
		return
	}
	// 验证是否为黑名单Token
	_, err := reCache.Get("blackToken", tokenStr)
	if err == nil {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: dataType.BlackToken, Message: "invalid token",
		})
		return
	}

	token, err := auth.CheckToken(tokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, dataType.JsonWrong{
			Code: dataType.WrongToken, Message: err.Error(),
		})
		return
	}
	c.Set("token", token)
}

func (l *loginStatus) WsHandle(w *wes.WContext) {
	if w.Conn.UserId == 0 || w.Conn.AuthExpireTime.Before(time.Now()) {
		w.Result(dataType.NoToken, "without auth")
	}
}

// NewLoginStatus token验证中间件，验证token是否正确、是否过期、是否已注销，并将token指针放入请求上下文中
func NewLoginStatus() Middleware {
	return &loginStatus{}
}

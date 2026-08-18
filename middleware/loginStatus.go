package middleware

import (
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type loginStatus struct {
}

// 登录验证中间件，将解析的token指针存入gin上下文
func (l *loginStatus) HttpHandle(c *gin.Context) {
	tokenStr := c.GetHeader("Token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dataType.JsonWrong{
			Code: dataType.NoToken, Message: "invalid token",
		})
		return
	}
	// 验证是否为黑名单Token
	_, err := reCache.Get("blackToken", tokenStr, nil)
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
		w.Result(dataType.WsAuthExpire, "auth expired")
		go w.Conn.Disconnect()
	}
}

// token 验证中间件实例
var AuthMiddle = &loginStatus{}

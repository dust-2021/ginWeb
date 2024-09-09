package middleware

import (
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
)

type Permission struct {
	permission []string
}

func containsAll(a, b []string) bool {

	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}

	for _, v := range b {
		if _, found := elementMap[v]; !found {
			return false
		}
	}
	return true
}

func (p Permission) Handle(c *gin.Context) {
	if len(p.permission) == 0 {
		c.Next()
		return
	}
	tokenStr, f := c.Get("token")
	if !f {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "without token",
		})
		return
	}
	token, flag := tokenStr.(auth.Token)
	if !flag {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "without token",
		})
		return
	}
	if !containsAll(token.Permission, p.permission) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "permission don't match",
		})
	}

}

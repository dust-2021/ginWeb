package middleware

import (
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
)

type Permission struct {
	Permission []string
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
	if len(p.Permission) == 0 {
		c.Next()
		return
	}
	tokenPtr, f := c.Get("token")
	if !f {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "without token",
		})
		return
	}
	token, flag := tokenPtr.(*auth.Token)
	if !flag {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "without token",
		})
		return
	}
	if !containsAll(token.Permission, p.Permission) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "Permission don't match",
		})
	}

}

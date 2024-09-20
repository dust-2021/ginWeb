package middleware

import (
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
)

// Permission 对比token中存储的权限是否足够
type Permission struct {
	// 必要权限
	Permission []string
	// 可选权限，只用包含每个子数组中的一个
	SelectPermission [][]string
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

func containOne(a []string, b [][]string) bool {
	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}
	for _, v := range b {
		flag := false
		for _, v1 := range v {
			if _, found := elementMap[v1]; found {
				flag = true
				break
			}
		}
		if !flag {
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
	if !containsAll(token.Permission, p.Permission) || !containOne(token.Permission, p.SelectPermission) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "Permission don't match",
		})
	}

}

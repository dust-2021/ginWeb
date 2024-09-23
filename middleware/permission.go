package middleware

import (
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
)

type permission struct {
	// 必要权限
	Permission []string
	// 可选权限，只用包含每个子数组中的一个
	SelectPermission [][]string
}

func containsAll(a, b []string) bool {
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
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
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
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
		// 某个多选一未通过
		if !flag {
			return false
		}
	}
	return true
}

func (p *permission) Handle(c *gin.Context) {
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
	if !(containsAll(token.Permission, p.Permission) && containOne(token.Permission, p.SelectPermission)) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "permission don't match",
		})
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

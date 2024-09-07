package middleware

import "github.com/gin-gonic/gin"

type Permission struct {
	permission []string
}

func (p Permission) Handle(c *gin.Context) error {
	if len(p.permission) == 0 {
		c.Next()
	}
	c.AbortWithStatus(403)
	return nil
}

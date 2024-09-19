package perm

import "github.com/gin-gonic/gin"

type Grant struct {
}

type data struct {
	Type int `json:"type" binding:"required"`
	ToId int `json:"toId" binding:"required"`
}

func (g Grant) GrantRole(ctx *gin.Context) {
	var data data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {

	}
}

func (g Grant) GrantGroup(ctx *gin.Context) {

}

func (g Grant) RegisterRoute(r string, router *gin.RouterGroup) {
	router.Handle("POST", r+"/role", g.GrantRole)
	router.Handle("POST", r+"/group", g.GrantGroup)
}

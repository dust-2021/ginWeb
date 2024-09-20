package perm

import (
	"ginWeb/middleware"
	"ginWeb/model/permissionMode"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/database"
	"github.com/gin-gonic/gin"
)

type Grant struct {
}

type data struct {
	ToId    int `json:"toId" binding:"required"`
	GrantId int `json:"grantId" binding:"required"`
}

// RoleToUser 赋予用户角色
func (g Grant) RoleToUser(ctx *gin.Context) {
	var data data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	var existed permissionMode.UserRole
	database.Db.Where("role_id = ? and user_id = ?", data.GrantId, data.ToId).First(&existed)
	if existed.Id != 0 {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 0, Message: "already existed",
		})
		return
	}
	var role permissionMode.Role
	var user systemMode.User
	resp := database.Db.Where("id = ?", data.GrantId).First(&role)
	if resp.Error != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "invalid role",
		})
		return
	}
	resp = database.Db.Where("id = ?", data.ToId).First(&user)
	if resp.Error != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "invalid user",
		})
		return
	}
	rec := permissionMode.UserRole{
		UserId: data.ToId,
		RoleId: data.GrantId,
	}
	resp = database.Db.Create(&rec)
	if resp.Error != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "failed: " + resp.Error.Error(),
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Data: "success",
	})
}

func (g Grant) GroupToUser(ctx *gin.Context) {}

func (g Grant) RegisterRoute(r string, router *gin.RouterGroup) {
	router.Handle("POST", r+"/roleToUser", middleware.Permission{Permission: []string{"admin"}}.Handle, g.RoleToUser)
	router.Handle("POST", r+"/groupToUser", g.GroupToUser)
}

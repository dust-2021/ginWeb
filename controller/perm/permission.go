package perm

import (
	"ginWeb/model/authMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/database"
	"github.com/gin-gonic/gin"
)

type Permission struct{}

type PermissionData struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
}

func (p Permission) Create(c *gin.Context) {
	var data PermissionData
	err := c.ShouldBindJSON(data)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongBody,
			Message: err.Error(),
		})
		return
	}
	per := authMode.Permissions{
		PermissionName: data.Name,
		Description:    data.Desc,
	}
	database.Db.Create(&per)
	c.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: "success",
	})
}
func (p Permission) Delete(c *gin.Context) {
	var data PermissionData
	err := c.ShouldBindJSON(data)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongBody,
			Message: err.Error(),
		})
		return
	}
	var per authMode.Permissions
	resp := database.Db.Table("permissions").Where("permission_name = ?", data.Name).First(&per)
	if resp.Error != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: resp.Error.Error(),
		})
		return
	}
	database.Db.Delete(&per)
	c.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: "success",
	})
}

func (p Permission) Update(c *gin.Context) {
}

func (p Permission) CreateGroup(c *gin.Context) {
}
func (p Permission) DeleteGroup(c *gin.Context) {
}
func (p Permission) CreateRole(c *gin.Context) {
}
func (p Permission) DeleteRole(c *gin.Context) {
}

func (p Permission) RegisterRoute(r string, g *gin.RouterGroup) {
	group := g.Group(r)
	group.Handle("GET", "create", p.Create)
	group.Handle("GET", "delete", p.Delete)
	group.Handle("POST", "update", p.Update)
	group.Handle("GET", "createGroup", p.CreateGroup)
	group.Handle("GET", "deleteGroup", p.DeleteGroup)
	group.Handle("GET", "createRole", p.CreateRole)
	group.Handle("GET", "deleteRole", p.DeleteRole)
}

package controller

import (
	"ginWeb/config"
	"ginWeb/middleware"
	"ginWeb/model"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Login struct {
}

type postData struct {
	Username string `json:"username" binding:"required,min=3,max=36"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

func (receiver Login) V1(c *gin.Context) {
	var j postData
	err := c.ShouldBindJSON(&j)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{
			Code: dataType.WrongBody, Message: err.Error(),
		})
		return
	}
	// 查找用户
	var u systemMode.User
	result := database.Db.Where("username = ?", j.Username).First(&u)
	pwd := auth.HashString(j.Password)
	if result.Error != nil || pwd != u.PasswordHash {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{Code: dataType.WrongData,
			Message: "username or password invalid"})
		return
	}
	// 查询权限
	permissions, err := model.GetPermissionById(u.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{Code: 1, Message: "get permission error: " + err.Error()})
		return
	}

	token := auth.Token{
		UserId:     u.Id,
		UserUUID:   u.Uuid,
		Username:   u.Username,
		Permission: permissions,
		Expire:     time.Now().Add(time.Second * time.Duration(config.Conf.Server.TokenExpire)),
	}
	data, err := token.Sign()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{
			Code:    dataType.Unknown,
			Message: "generate token failed",
		})
		return
	}

	c.JSON(200, dataType.JsonRes{Code: dataType.Success, Data: data})
}

func (receiver Login) RegisterRoute(r string, g *gin.RouterGroup) {
	// 添加独立限流器
	g.Handle("POST", r, middleware.NewIpLimiter(5, 0, 0, g.BasePath()+r).HttpHandle, receiver.V1)
}

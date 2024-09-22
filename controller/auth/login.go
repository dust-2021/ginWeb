package controller

import (
	"ginWeb/config"
	"ginWeb/middleware"
	"ginWeb/model"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Login struct {
}

type postData struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

func (receiver Login) V1(c *gin.Context) {
	var j postData
	err := c.BindJSON(&j)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	// 查找用户
	var u systemMode.User
	result := database.Db.Where("phone = ?", j.Username).Or("email = ?", j.Username).First(&u)
	pwd, err := auth.HashPassword(j.Password)
	if result.Error != nil || err != nil || pwd != u.PasswordHash {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{Code: 1,
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
		Permission: permissions,
		Expire:     time.Now().Add(time.Second * time.Duration(config.Conf.Server.TokenExpire)),
	}
	data := ""
	if config.Conf.Server.TokenEncrypt {
		data, err = token.AesEncrypt()
	} else {
		data, err = token.Sign()
	}
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    1,
			Message: "generate token failed",
		})
		return
	}

	c.JSON(200, dataType.JsonRes{Code: 0, Data: data})
}

func (receiver Login) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("POST", r, middleware.NewRouteLimiter(5, 0, 0).Handle, receiver.V1)
}

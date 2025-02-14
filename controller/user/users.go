package user

import (
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"github.com/gin-gonic/gin"
)

type Users struct {
}

type data struct {
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (u Users) Create(ctx *gin.Context) {
	var data data
	err := ctx.ShouldBind(&data)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: err.Error(),
		})
		return
	}

	hashed, err := auth.HashPassword(data.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: err.Error(),
		})
		return
	}
	newUser := systemMode.User{
		Phone:        data.Phone,
		Email:        data.Email,
		PasswordHash: hashed,
		Available:    true,
	}
	flag, err := newUser.Exist()
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.Unknown,
			Message: err.Error(),
		})
		return
	}
	if flag {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.AlreadyExist,
			Message: "User already exist",
		})
		return
	}
	_ = database.Db.Create(&newUser)
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
	})
}

func (u Users) RegisterRoute(r string, router *gin.RouterGroup) {
	g := router.Group(r)
	g.Handle("POST", "create", u.Create)
}

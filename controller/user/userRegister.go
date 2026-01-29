package user

import (
	"ginWeb/config"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type dataCreate struct {
	UserName string `json:"username" binding:"required,max=32,min=3"`
	Password string `json:"password" binding:"required,max=32,min=6"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func PublicRegister(ctx *gin.Context) {
	if !config.DynamicConf.EnablePublicRegister() {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.Forbidden,
			Message: "public register is disabled",
		})
		return
	}
	var reqData dataCreate
	err := ctx.ShouldBindJSON(&reqData)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: err.Error(),
		})
		return
	}
	//if reqData.Phone == "" && reqData.Email == "" {
	//	ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
	//		Code:    dataType.WrongData,
	//		Message: "phone or email is required",
	//	})
	//}
	// TODO 按需添加邮箱或手机验证
	hashed := auth.HashString(reqData.Password)
	newUser := systemMode.User{
		Uuid:         uuid.New().String(),
		Phone:        reqData.Phone,
		Email:        reqData.Email,
		Username:     reqData.UserName,
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

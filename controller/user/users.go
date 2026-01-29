package user

import (
	"ginWeb/middleware"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"ginWeb/utils/tools"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	nanoid "github.com/matoous/go-nanoid/v2"
)

type Users struct {
}

type dataUpdate struct {
	Password *string `json:"password" binding:"max=32,min=3"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
}

// Update 更新账号信息
func (u Users) Update(ctx *gin.Context) {
	var reqData dataUpdate
	err := ctx.ShouldBindJSON(&reqData)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: err.Error(),
		})
		return
	}
	token, flag := ctx.Get("token")
	if !flag {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: "token required",
		})
		return
	}
	t, ok := token.(*auth.Token)
	if !ok {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: "token required",
		})
		return
	}
	oldUser, err := systemMode.GetUserByID(t.UserId)
	if err != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.NotFound,
			Message: "user not found",
		})
		return
	}

	if reqData.Password != nil {
		oldUser.PasswordHash = auth.HashString(*reqData.Password)
	}
	if reqData.Phone != nil {
		oldUser.Phone = *reqData.Phone
	}
	if reqData.Email != nil {
		oldUser.Email = *reqData.Email
	}
	result := database.Db.Table("user").Updates(oldUser)
	if result.Error != nil {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.Unknown,
			Message: result.Error.Error(),
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
	})
}

type createPiecesResp struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreatePieces 批量创建账号
func (u Users) CreatePieces(ctx *gin.Context) {
	count := ctx.Query("count")
	c, err := strconv.Atoi(count)
	if err != nil || c > 20 {
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: "wrong count",
		})
		return
	}
	var result []createPiecesResp
	tx := database.Db.Begin()
	for range c {
		randomName, _ := nanoid.New(12)
		username := uuid.New().String()
		password := tools.GenerateRandomString(6)
		resp := tx.Table("user").Create(&systemMode.User{
			Uuid:         username,
			Username:     randomName,
			PasswordHash: auth.HashString(password),
			Available:    true,
		})
		result = append(result, createPiecesResp{
			Username: randomName,
			Password: password,
		})
		if resp.Error != nil {
			ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
				Code:    dataType.Unknown,
				Message: resp.Error.Error(),
			})
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: result,
	})
}

func (u Users) RegisterRoute(r string, router *gin.RouterGroup) {
	g := router.Group(r)
	g.Handle("POST", "update", u.Update)
	g.Handle("GET", "createPieces", middleware.NewPermission([]string{"admin"}).HttpHandle, u.CreatePieces)
}

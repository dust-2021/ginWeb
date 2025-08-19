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
)

type Users struct {
}

type dataCreate struct {
	UserName string `json:"username" binding:"required,max=32,min=3"`
	Password string `json:"password" binding:"required,max=32,min=6"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (u Users) Create(ctx *gin.Context) {
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
	var result []interface{}
	tx := database.Db.Begin()
	for i := 0; i < c; i++ {
		username := uuid.New().String()
		password := tools.GenerateRandomString(6)
		resp := tx.Table("user").Create(&systemMode.User{
			Uuid:         username,
			Username:     username,
			PasswordHash: auth.HashString(password),
			Available:    true,
		})
		result = append(result, map[string]interface{}{
			"username": username,
			"password": password,
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
	g.Handle("POST", "create", u.Create)
	g.Handle("POST", "update", u.Update)
	g.Handle("GET", "createPieces", middleware.NewPermission([]string{"admin"}).HttpHandle, u.CreatePieces)
}

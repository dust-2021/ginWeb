package inital

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/model"
	"ginWeb/model/authMode"
	"ginWeb/model/systemMode"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"ginWeb/utils/loguru"
)

// InitializeMode 初始化数据表，创建管理员账号和权限
func InitializeMode() {
	db := database.Db
	// 设置字符集，默认字符集不是utf8
	err := db.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(&systemMode.User{}, &authMode.Permissions{},
		&authMode.Group{}, &authMode.UserGroup{}, &authMode.GroupPermission{},
		&authMode.Role{}, &authMode.UserRole{}, &authMode.RolePermission{},
	)
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "SYSTEM", fmt.Sprintf("create table failed %s", err.Error()))
	}
	hashPwd := auth.HashPassword(config.Conf.Server.AdminUser.Password)
	user := systemMode.User{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		Username:     config.Conf.Server.AdminUser.Username,
		Phone:        config.Conf.Server.AdminUser.Phone,
		Email:        config.Conf.Server.AdminUser.Email,
		PasswordHash: hashPwd,
	}
	role := authMode.Role{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		RoleName:    "admin",
		Description: "系统管理员",
	}
	perm := authMode.Permissions{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		PermissionName: "admin",
		Description:    "管理员权限",
	}
	resp := db.Create(&user)
	if resp.Error != nil {
		loguru.SimpleLog(loguru.Warn, "SYSTEM", fmt.Sprintf("init admin user failed %s", resp.Error.Error()))
	}
	resp = db.Create(&role)
	if resp.Error != nil {
		loguru.SimpleLog(loguru.Warn, "SYSTEM", fmt.Sprintf("init admin role failed %s", resp.Error.Error()))
	}
	resp = db.Create(&perm)
	if resp.Error != nil {
		loguru.SimpleLog(loguru.Warn, "SYSTEM", fmt.Sprintf("init admin user permission %s", resp.Error.Error()))
	}
	if err != nil {
		return
	}
}

package inital

import (
	"ginWeb/config"
	"ginWeb/model"
	"ginWeb/model/permissionMode"
	"ginWeb/model/systemMode"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"ginWeb/utils/loguru"
)

// InitializeMode 初始化数据表，创建管理员账号和权限
func InitializeMode() {
	db := database.Db
	err := db.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(&systemMode.User{}, &permissionMode.Permissions{},
		&permissionMode.Group{}, &permissionMode.UserGroup{}, &permissionMode.GroupPermission{},
		&permissionMode.Role{}, &permissionMode.UserRole{}, &permissionMode.RolePermission{},
	)
	if err != nil {
		loguru.Logger.Fatalf("create table failed %s", err.Error())
	}
	hashPwd, err := auth.HashPassword(config.Conf.Server.AdminUser.Password)
	if err != nil {
		loguru.Logger.Fatalf("init admin user failed %s", err.Error())
	}
	user := systemMode.User{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		Phone:        config.Conf.Server.AdminUser.Phone,
		Email:        config.Conf.Server.AdminUser.Email,
		PasswordHash: hashPwd,
	}
	role := permissionMode.Role{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		RoleName:    "admin",
		Description: "系统管理员",
	}
	perm := permissionMode.Permissions{
		BaseModel: model.BaseModel{
			Id: 1,
		},
		PermissionName: "admin",
		Description:    "管理员权限",
	}
	resp := db.Create(&user)
	if resp.Error != nil {
		loguru.Logger.Warnf("init admin user failed %s", resp.Error.Error())
	}
	resp = db.Create(&role)
	if resp.Error != nil {
		loguru.Logger.Warnf("init admin role failed %s", resp.Error.Error())
	}
	resp = db.Create(&perm)
	if resp.Error != nil {
		loguru.Logger.Warnf("init admin permission failed %s", resp.Error.Error())
	}
	if err != nil {
		return
	}
}

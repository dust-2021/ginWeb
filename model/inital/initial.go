package inital

import (
	"ginWeb/model/permissionMode"
	"ginWeb/model/systemMode"
	"ginWeb/utils/database"
)

func InitializeMode() {
	db := database.Db
	err := db.AutoMigrate(&systemMode.User{}, &permissionMode.Permissions{},
		&permissionMode.Group{}, &permissionMode.UserGroup{}, &permissionMode.GroupPermission{},
		&permissionMode.Role{}, &permissionMode.UserRole{}, &permissionMode.RolePermission{},
	)
	if err != nil {
		return
	}
}

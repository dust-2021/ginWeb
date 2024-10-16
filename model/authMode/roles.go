package authMode

import "ginWeb/model"

type Role struct {
	model.BaseModel `gorm:"embedded"`
	RoleName        string `gorm:"size:20;NOT NULL"`
	Description     string `gorm:"size:255"`
}

type UserRole struct {
	model.BaseModel `gorm:"embedded"`
	RoleId          int `gorm:"index;NOT NULL"`
	UserId          int `gorm:"index;NOT NULL"`
}

type RolePermission struct {
	model.BaseModel `gorm:"embedded"`
	RoleId          int `gorm:"index;NOT NULL"`
	PermissionId    int `gorm:"index;NOT NULL"`
}

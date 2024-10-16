package authMode

import "ginWeb/model"

type Group struct {
	model.BaseModel `gorm:"embedded"`
	GroupName       string `gorm:"size:20;unique;not null"`
	Description     string `gorm:"size:255;"`
}

type UserGroup struct {
	model.BaseModel `gorm:"embedded"`
	GroupId         int `gorm:"index;not null"`
	UserId          int `gorm:"index;not null"`
}

type GroupPermission struct {
	model.BaseModel `gorm:"embedded"`
	GroupId         int `gorm:"index;not null"`
	PermissionId    int `gorm:"index;not null"`
}

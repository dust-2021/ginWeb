package authMode

import "ginWeb/model"

type Permissions struct {
	model.BaseModel `gorm:"embedded"`
	PermissionName  string `gorm:"type:varchar(20);not null"`
	Description     string `gorm:"type:varchar(255);default:''"`
}

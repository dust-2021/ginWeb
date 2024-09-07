package permissionMode

import "ginWeb/model"

type Permissions struct {
	model.BaseModel
	PermissionName string `gorm:"type:varchar(20);not null"`
	Description    string `gorm:"type:varchar(255);default:''"`
}

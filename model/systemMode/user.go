package systemMode

import (
	"ginWeb/model"
)

type User struct {
	model.BaseModel `gorm:"embedded"`
	Uuid            string `gorm:"size:36;NOT NULL;UNIQUE"`
	Phone           string `gorm:"size:20;NOT NULL;DEFAULT:''"`
	Email           string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	PasswordHash    string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	Available       bool   `gorm:"DEFAULT:false"`
}

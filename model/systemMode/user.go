package systemMode

import (
	"ginWeb/model"
	"ginWeb/utils/database"
)

type User struct {
	model.BaseModel `gorm:"embedded"`
	Uuid            string `gorm:"size:36;NOT NULL;UNIQUE"`
	Phone           string `gorm:"size:20;NOT NULL;DEFAULT:''"`
	Email           string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	PasswordHash    string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	Available       bool   `gorm:"DEFAULT:false"`
}

func (u User) Exist() (bool, error) {
	var c int64
	result := database.Db.Table("users").Where("phone = ?", u.Phone).Or("email = ?", u.Email).Count(&c)
	if result.Error != nil {
		return false, result.Error
	}
	return c != 0, nil
}

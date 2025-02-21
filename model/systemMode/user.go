package systemMode

import (
	"errors"
	"ginWeb/model"
	"ginWeb/utils/database"
)

type User struct {
	model.BaseModel `gorm:"embedded"`
	Uuid            string `gorm:"size:36;NOT NULL;UNIQUE"`
	Username        string `gorm:"size:100;NOT NULL;UNIQUE"`
	Phone           string `gorm:"size:20;NOT NULL;DEFAULT:''"`
	Email           string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	PasswordHash    string `gorm:"size:255;NOT NULL;DEFAULT:''"`
	Available       bool   `gorm:"DEFAULT:false"`
}

// Exist 判断是否存在
func (u *User) Exist() (bool, error) {
	var c int64
	if u.Username == "" {
		return false, errors.New("blank username")
	}
	result := database.Db.Table("user").Where("username = ?", u.Username).Count(&c)
	if result.Error != nil {
		return false, result.Error
	}
	return c != 0, nil
}

// GetUserByID 根据Id获取user信息
func GetUserByID(id int64) (*User, error) {
	var user User
	result := database.Db.Table("user").Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func GetUserByUsername(username string) (*User, error) {
	var user User
	result := database.Db.Table("user").Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

package model

import (
	"errors"
	"ginWeb/utils/database"
)

type BaseModel struct {
	Id        int64 `gorm:"primary_key;AUTO_INCREMENT"`
	CreatedAt int64 `gorm:"autoCreateTime:milli;index"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
	Deleted   bool  `gorm:"type:boolean;default:false"`
}

func (m BaseModel) Delete() error {
	if m.Id == 0 {
		return errors.New("delete failed without id")
	}
	database.Db.Find("id = ?", m.Id).Update("deleted", true)
	return nil
}

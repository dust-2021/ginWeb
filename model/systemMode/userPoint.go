package systemMode

import "ginWeb/model"

type UserPoint struct {
	model.BaseModel `gorm:"embedded"`
	Points          int `gorm:"DEFAULT:0"`
}

func (u UserPoint) Decrease(cost int64) error {

	return nil
}

type UserLabel struct {
	model.BaseModel `gorm:"embedded"`
}

package model

type BaseModel struct {
	Id        int64 `gorm:"primary_key;AUTO_INCREMENT"`
	CreatedAt int64 `gorm:"autoCreateTime:milli;index"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
	Deleted   bool  `gorm:"type:boolean;default:false"`
}

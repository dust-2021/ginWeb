package exchangeMode

import (
	"ginWeb/model"
	"github.com/shopspring/decimal"
)

type SpotLog struct {
	model.BaseModel `gorm:"embedded"`
	Type            string          `gorm:"size:20;not null"`
	Side            string          `gorm:"size:10;not null"`
	Quantity        decimal.Decimal `gorm:"type:decimal(20,8)"`
	Price           decimal.Decimal `gorm:"type:decimal(20,8)"`
	Fake            bool            `gorm:""`
}

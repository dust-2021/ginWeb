package exchangeCore

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

// TradeArgs 交易参数
type TradeArgs struct {
	Symbol       string          `json:"symbol"`
	Side         TradeSide       `json:"side"`
	Type         TradeType       `json:"type"`
	Quantity     decimal.Decimal `json:"quantity"`
	PositionSide PositionType    `json:"positionSide"`
	Price        decimal.Decimal `json:"price"`
	ReduceOnly   bool            `json:"reduceOnly"`
	TimeInForce  TradeStrategy   `json:"timeInForce"`
}

// ExAccount 账户
type ExAccount struct {
	Config      *CoreConfig
	Balance     map[string]*Balance
	SpotBalance map[string]*Balance
	SwapBalance map[string]*SwapBalance
	Position    map[string]map[PositionType]*SwapPosition
}

// SetLeverage 修改杠杆
func (e *ExAccount) SetLeverage(symbol string, leverage uint8) error {
	pos, f := e.Position[symbol]
	// 未初始化
	if !f {
		e.Config.Leverage[symbol] = leverage
		return nil
	}
	withPos := false
	for _, v := range pos {
		if !v.Quantity.IsZero() {
			withPos = true
			break
		}
	}
	// 无仓位
	if !withPos {
		e.Config.Leverage[symbol] = leverage
		return nil
	}

	if leverage < e.Config.Leverage[symbol] {
		return fmt.Errorf("leverage must be greater than the current leverage")
	} else if leverage == e.Config.Leverage[symbol] {
		return nil
	}

	// TODO 存在仓位时上调杠杆
	return errors.New("not support set leverage with pos exist yet")
}

// Transfer 划转 仅支持资金、现货和合约之间划转
func (e *ExAccount) Transfer(symbol string, type_ TransferType, amount decimal.Decimal) error {
	var from BaseBalance
	var to BaseBalance
	switch type_ {
	case MainBalance:
		from, _ = e.SpotBalance[symbol]
		to, _ = e.Balance[symbol]
	case MainFuture:
		from, _ = e.SpotBalance[symbol]
		to, _ = e.SwapBalance[symbol]
	case BalanceMain:
		from, _ = e.Balance[symbol]
		to, _ = e.SpotBalance[symbol]
	case BalanceFuture:
		from, _ = e.Balance[symbol]
		to, _ = e.SwapBalance[symbol]
	case FutureMain:
		from, _ = e.SwapBalance[symbol]
		to, _ = e.SpotBalance[symbol]
	case FutureBalance:
		from, _ = e.SwapBalance[symbol]
		to, _ = e.SpotBalance[symbol]
	default:
		return errors.New("not support transfer type")
	}
	flag := from.Sub(amount)
	if !flag {
		return errors.New("transfer amount not enough")
	}
	to.Add(amount)
	return nil
}

func (e *ExAccount) Cancellation() (err error) {
	return nil
}

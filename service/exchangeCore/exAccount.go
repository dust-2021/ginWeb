package exchangeCore

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

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
func (receiver *ExAccount) SetLeverage(symbol string, leverage uint8) error {
	pos, f := receiver.Position[symbol]
	// 未初始化
	if !f {
		receiver.Config.Leverage[symbol] = leverage
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
		receiver.Config.Leverage[symbol] = leverage
		return nil
	}

	if leverage < receiver.Config.Leverage[symbol] {
		return fmt.Errorf("leverage must be greater than the current leverage")
	} else if leverage == receiver.Config.Leverage[symbol] {
		return nil
	}

	// TODO 存在仓位时上调杠杆
	return errors.New("not support set leverage with pos exist yet")
}

// Transfer 划转 仅支持资金、现货和合约之间划转
func (receiver *ExAccount) Transfer(symbol string, type_ TransferType, amount decimal.Decimal) error {
	var from BaseBalance
	var to BaseBalance
	switch type_ {
	case MainBalance:
		from, _ = receiver.SpotBalance[symbol]
		to, _ = receiver.Balance[symbol]
	case MainFuture:
		from, _ = receiver.SpotBalance[symbol]
		to, _ = receiver.SwapBalance[symbol]
	case BalanceMain:
		from, _ = receiver.Balance[symbol]
		to, _ = receiver.SpotBalance[symbol]
	case BalanceFuture:
		from, _ = receiver.Balance[symbol]
		to, _ = receiver.SwapBalance[symbol]
	case FutureMain:
		from, _ = receiver.SwapBalance[symbol]
		to, _ = receiver.SpotBalance[symbol]
	case FutureBalance:
		from, _ = receiver.SwapBalance[symbol]
		to, _ = receiver.SpotBalance[symbol]
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

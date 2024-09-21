package exchangeCore

import "github.com/shopspring/decimal"

// PositionType 持仓类型
type PositionType uint8

// TransferType 划转类型
type TransferType uint8

type TradeSide uint8

type TradeType uint8

type TradeStrategy uint8

// 持仓方向
const (
	Both PositionType = iota
	Short
	Long
)

// 划转类型
const (
	MainFuture TransferType = iota
	FutureMain
	MainBalance
	BalanceMain
	FutureBalance
	BalanceFuture
)

// 交易方向
const (
	BUY TradeSide = iota
	SELL
)

// 交易方式
const (
	MARKET TradeType = iota
	LIMIT
)

// 交易策略
const (
	GTC TradeStrategy = iota
	OTC
)

func (receiver PositionType) String() string {
	switch receiver {
	case Both:
		return "BOTH"
	case Short:
		return "SHORT"
	case Long:
		return "LONG"
	default:
		return "unknown"
	}
}

func (receiver TransferType) String() string {
	switch receiver {
	case MainBalance:
		return "MAIN_FUNDING"
	case MainFuture:
		return "MAIN_UMFUTURE"
	case FutureMain:
		return "UMFUTURE_MAIN"
	case FutureBalance:
		return "UMFUTURE_FUNDING"
	case BalanceMain:
		return "FUNDING_MAIN"
	case BalanceFuture:
		return "FUNDING_UMFUTURE"
	default:
		return "unknown"
	}
}

func (t TradeSide) String() string {
	switch t {
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	default:
		return "unknown"
	}
}

func (t TradeType) String() string {
	switch t {
	case MARKET:
		return "MARKET"
	case LIMIT:
		return "LIMIT"
	default:
		return "unknown"
	}
}

func (t TradeStrategy) String() string {
	switch t {
	case GTC:
		return "GTC"
	default:
		return "unknown"
	}
}

// CoreConfig 账户配置
type CoreConfig struct {
	MarginType   map[string]bool
	Leverage     map[string]uint8
	Other        map[string]interface{}
	PositionSide map[string]bool
}

type BaseBalance interface {
	Add(amount decimal.Decimal) bool
	Sub(amount decimal.Decimal) bool
}

// Balance 现货资金
type Balance struct {
	Symbol   string
	Quantity decimal.Decimal
}

func (b *Balance) Add(amount decimal.Decimal) bool {
	b.Quantity = b.Quantity.Add(amount)
	return true
}
func (b *Balance) Sub(amount decimal.Decimal) bool {
	if amount.Cmp(b.Quantity) < 0 {
		return false
	}
	b.Quantity = b.Quantity.Sub(amount)
	return true
}

// SwapBalance 合约资金
type SwapBalance struct {
	Symbol      string          `json:"symbol"`
	Quantity    decimal.Decimal `json:"quantity"`
	Freeze      decimal.Decimal `json:"freeze"`
	MaxWithdraw decimal.Decimal `json:"maxWithdraw"`
}

func (receiver *SwapBalance) Sub(quantity decimal.Decimal) bool {
	if receiver.Quantity.LessThan(quantity) || receiver.MaxWithdraw.LessThan(quantity) {
		return false
	}
	receiver.Quantity.Sub(quantity)
	receiver.MaxWithdraw.Sub(quantity)
	return true
}

func (receiver *SwapBalance) Add(quantity decimal.Decimal) bool {
	receiver.Quantity.Add(quantity)
	receiver.MaxWithdraw.Add(quantity)
	return true
}

// SwapPosition 合约仓位
type SwapPosition struct {
	Symbol          string          `json:"symbol"`
	Quantity        decimal.Decimal `json:"quantity"`
	BreakEventPrice decimal.Decimal `json:"breakEventPrice"`
	PositionSide    PositionType    `json:"positionSide"`
}

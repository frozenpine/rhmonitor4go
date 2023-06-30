package sinker

import (
	"time"

	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/pkg/errors"
)

type MsgFormat uint8

func (msgFmt *MsgFormat) Set(value string) error {
	switch value {
	case "proto3":
		*msgFmt = MsgProto3
	case "json":
		*msgFmt = MsgJson
	case "msgpack":
		*msgFmt = MsgPack
	default:
		return errors.New("invalid msg format")
	}

	return nil
}

//go:generate stringer -type MsgFormat -linecomment
const (
	MsgProto3 MsgFormat = iota // proto3
	MsgJson                    // json
	MsgPack                    // msgpack
)

type SinkAccount struct {
	InvestorID string    `sql:"account_id" json:"account_id" msgpack:"account_id"`
	TradingDay string    `sql:"trading_day" json:"trading_day" msgpack:"trading_day"`
	Timestamp  time.Time `sql:"timestamp" json:"timestamp" msgpack:"timestamp"`
	PreBalance float64   `sql:"pre_balance" json:"pre_balance" msgpack:"pre_balance"`
	Balance    float64   `sql:"balance" json:"balance" msgpack:"balance"`
	Deposit    float64   `sql:"deposit" json:"deposit" msgpack:"deposit"`
	Withdraw   float64   `sql:"withdraw" json:"withdraw" msgpack:"withdraw"`
	Profit     float64   `sql:"profit" json:"profit" msgpack:"profit"`
	Fee        float64   `sql:"fee" json:"fee" msgpack:"fee"`
	Margin     float64   `sql:"margin" json:"margin" msgpack:"margin"`
	Available  float64   `sql:"available" json:"available" msgpack:"available"`
}

func (acct *SinkAccount) FromAccount(value *service.Account) {
	acct.TradingDay = value.TradingDay
	acct.InvestorID = value.Investor.InvestorId
	acct.Timestamp = time.UnixMilli(value.Timestamp)
	acct.PreBalance = value.PreBalance
	acct.Deposit = value.Deposit
	acct.Withdraw = value.Withdraw
	acct.Profit = value.CloseProfit + value.PositionProfit
	acct.Fee = value.Commission + value.FrozenCommission
	acct.Margin = value.CurrentMargin + value.FrozenMargin
	acct.Available = value.Available
	acct.Balance = service.RohonCaculateDynamicBalance(
		value.PreBalance, value.Deposit, value.Withdraw,
		value.CloseProfit, value.PositionProfit, value.Commission,
	)
}

type BarMode uint8

const (
	Continuous BarMode = 1 << iota
	FirstTick
)

type SinkAccountBar struct {
	TradingDay string    `sql:"trading_day" json:"trading_day" msgpack:"trading_day"`
	AccountID  string    `sql:"account_id" json:"account_id" msgpack:"account_id"`
	Timestamp  time.Time `sql:"timestamp" json:"timestamp" msgpack:"timestamp"`
	Duration   string    `sql:"duration" json:"duration" msgpack:"duration"`
	Open       float64   `sql:"open" json:"open" msgpack:"open"`
	Close      float64   `sql:"close" json:"close" msgpack:"close"`
	Highest    float64   `sql:"high" json:"high" msgpack:"high"`
	Lowest     float64   `sql:"low" json:"low" msgpack:"low"`
}

type SinkAccountPosition struct {
	TradingDay   string  `sql:"trading_day" json:"trading_day" msgpack:"trading_day"`
	AccountID    string  `sql:"account_id" json:"account_id" msgpack:"account_id"`
	ProductID    string  `sql:"priduct_id" json:"product_id" msgpack:"product_id"`
	InstrumentID string  `sql:"instrument_id" json:"instrument_id" msgpack:"instrument_id"`
	HedgeFlag    string  `sql:"hedge_flag" json:"hedge_flag" msgpack:"hedge_flag"`
	Diretion     string  `sql:"direction" json:"direction" msgpack:"direction"`
	VolumeTotal  int     `sql:"volume_total" json:"volume_total" msgpack:"volume_total"`
	Margin       float64 `sql:"margin" json:"margin" msgpack:"margin"`
	AvgOpenPrice float64 `sql:"avg_open_price" json:"avg_open_price" msgpack:"avg_open_price"`
	AvgPosPrice  float64 `sql:"avg_pos_price" json:"avg_pos_price" msgpack:"avg_pos_price"`
	VolumeToday  int     `sql:"volume_today" json:"volume_today" msgpack:"volume_today"`
	FrozenVolume int     `sql:"frozen_volume" json:"frozen_volume" msgpack:"frozen_volume"`
}

func (pos *SinkAccountPosition) FromPosition(value *service.Position, trading_day string) {
	pos.TradingDay = trading_day
	pos.AccountID = value.Investor.GetInvestorId()
	pos.ProductID = value.GetProductId()
	pos.InstrumentID = value.InstrumentId
	pos.HedgeFlag = value.HedgeFlag.String()
	pos.Diretion = value.Direction.String()
	pos.VolumeTotal = int(value.Volume)
	pos.Margin = value.Margin
	pos.AvgOpenPrice = value.AvgOpenPriceByVol
	pos.AvgPosPrice = value.AvgOpenPrice
	pos.VolumeToday = int(value.TodayVolume)
	pos.FrozenVolume = int(value.FrozenVolume)
}

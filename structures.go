package rhmonitor4go

import (
	"bytes"
	"strconv"
	"sync"
)

var stringBuffer = sync.Pool{New: func() any { return make([]byte, 0, 100) }}

type StringBuffer struct {
	buffer *bytes.Buffer
	under  []byte
}

func (buf *StringBuffer) Release() {
	stringBuffer.Put(buf.under[:0])
}

func (buf *StringBuffer) WriteString(v string) (int, error) {
	return buf.buffer.WriteString(v)
}

func (buf *StringBuffer) String() string {
	return buf.buffer.String()
}

func NewStringBuffer() *StringBuffer {
	under := stringBuffer.Get().([]byte)

	buf := StringBuffer{
		buffer: bytes.NewBuffer(under),
		under:  under,
	}

	return &buf
}

type RspInfo struct {
	ErrorID  int
	ErrorMsg string
}

func (rsp *RspInfo) Error() string {
	buff := bytes.NewBuffer(nil)

	buff.WriteString("[" + strconv.Itoa(rsp.ErrorID) + "] ")
	buff.WriteString(rsp.ErrorMsg)

	return buff.String()
}

type RiskUser struct {
	UserID     string
	Password   string
	MACAddress string
}

func (usr RiskUser) IsValid() bool {
	return usr.UserID != "" && usr.Password != ""
}

type RspUserLogin struct {
	UserID            string
	PrivilegeType     PrivilegeType
	InfoPrivilegeType InfoPrivilegeType
	TradingDay        string
	LoginTime         string
}

type RspUserLogout struct {
	UserID string
}

type Investor struct {
	BrokerID   string
	InvestorID string
}

func (i Investor) Identity() string {
	out := NewStringBuffer()
	defer out.Release()

	if i.BrokerID != "" {
		out.WriteString(i.BrokerID)
		out.WriteString(".")
	}

	out.WriteString(i.InvestorID)

	return out.String()
}

type Account struct {
	//经纪公司代码
	BrokerID string
	//投资者帐号
	AccountID string
	//上次质押金额
	PreMortgage float64
	//上次信用额度
	PreCredit float64
	//上次存款额
	PreDeposit float64
	//上次结算准备金
	PreBalance float64
	//上次占用的保证金
	PreMargin float64
	//利息基数
	InterestBase float64
	//利息收入
	Interest float64
	//入金金额
	Deposit float64
	//出金金额
	Withdraw float64
	//冻结的保证金
	FrozenMargin float64
	//冻结的资金
	FrozenCash float64
	//冻结的手续费
	FrozenCommission float64
	//当前保证金总额
	CurrMargin float64
	//资金差额
	CashIn float64
	//手续费
	Commission float64
	//平仓盈亏
	CloseProfit float64
	//持仓盈亏
	PositionProfit float64
	//期货结算准备金
	Balance float64
	//可用资金
	Available float64
	//可取资金
	WithdrawQuota float64
	//基本准备金
	Reserve float64
	//交易日
	TradingDay string
	//结算编号
	SettlementID int
	//信用额度
	Credit float64
	//质押金额
	Mortgage float64
	//交易所保证金
	ExchangeMargin float64
	//投资者交割保证金
	DeliveryMargin float64
	//交易所交割保证金
	ExchangeDeliveryMargin float64
	//保底期货结算准备金
	ReserveBalance float64
	//币种代码
	CurrencyID string
	//上次货币质入金额
	PreFundMortgageIn float64
	//上次货币质出金额
	PreFundMortgageOut float64
	//货币质入金额
	FundMortgageIn float64
	//货币质出金额
	FundMortgageOut float64
	//货币质押余额
	FundMortgageAvailable float64
	//可质押货币金额
	MortgageableFund float64
	//特殊产品占用保证金
	SpecProductMargin float64
	//特殊产品冻结保证金
	SpecProductFrozenMargin float64
	//特殊产品手续费
	SpecProductCommission float64
	//特殊产品冻结手续费
	SpecProductFrozenCommission float64
	//特殊产品持仓盈亏
	SpecProductPositionProfit float64
	//特殊产品平仓盈亏
	SpecProductCloseProfit float64
	//根据持仓盈亏算法计算的特殊产品持仓盈亏
	SpecProductPositionProfitByAlg float64
	//特殊产品交易所保证金
	SpecProductExchangeMargin float64
	//业务类型
	BizType BusinessType
	//延时换汇冻结金额
	FrozenSwap float64
	//剩余换汇额度
	RemainSwap float64
	//证券持仓市值
	TotalStockMarketValue float64
	//期权持仓市值
	TotalOptionMarketValue float64
	//动态权益
	DynamicMoney float64
	//权利金收支
	Premium float64
	//市值权益
	MarketValueEquity float64
}

func (acct Account) Identity() string {
	out := NewStringBuffer()
	defer out.Release()

	if acct.BrokerID != "" {
		out.WriteString(acct.BrokerID)
		out.WriteString(".")
	}

	out.WriteString(acct.AccountID)

	return out.String()
}

type Position struct {
	//投资者代码
	InvestorID string
	//经纪公司代码
	BrokerID string
	//合约类别
	ProductID string
	//合约代码
	InstrumentID string
	//投机套保标志
	HedgeFlag HedgeFlag
	//持仓方向
	Direction Direction
	//持仓数量
	Volume int
	//持仓保证金
	Margin float64
	//逐笔开仓均价
	AvgOpenPriceByVol float64
	//逐日开仓均价
	AvgOpenPrice float64
	//今仓数量
	TodayVolume int
	//冻结持仓数量
	FrozenVolume int
	//信息类型
	EntryType uint8
}

func (pos *Position) Identity() string {
	out := NewStringBuffer()
	defer out.Release()

	if pos.BrokerID != "" {
		out.WriteString(pos.BrokerID)
		out.WriteString(".")
	}
	out.WriteString(pos.InvestorID)

	return out.String()
}

type OffsetOrder struct {
	//投资者
	InvestorID string
	//经纪公司代码
	BrokerID string
	//合约ID
	InstrumentID string
	//方向
	Direction Direction
	//手数
	Volume int
	//价格
	Price float64
	//组合开平标志
	ComboOffsetFlag [5]byte
	//组合投机套保标志
	ComboHedgeFlag [5]byte
}

type Order struct {
	BrokerID             string
	InvestorID           string
	InstrumentID         string
	OrderRef             string
	UserID               string
	PriceType            OrderPriceType
	Direction            Direction
	ComboOffsetFlag      [5]byte
	ComboHedgeFlag       [5]byte
	LimitPrice           float64
	VolumeTotalOriginal  int
	TimeCondition        TimeCondition
	GTDDate              string
	VolumeCondition      VolumeCondition
	MinVolume            int
	ContingentCondition  ContingentCondition
	StopPrice            float64
	ForceCloseReason     ForceCloseReason
	IsAutoSuspend        bool
	BusinessUnit         string
	RequestID            int
	OrderLocalID         string
	ExchangeID           string
	ParticipantID        string
	ClientID             string
	ExchangeInstID       string
	TraderID             string
	InstallID            int
	OrderSubmitStatus    OrderSubmitStatus
	NotifySequence       int
	TradingDay           string
	SettlementID         int
	OrderSysID           string
	OrderSource          OrderSource
	OrderStatus          OrderStatus
	OrderType            OrderType
	VolumeTraded         int
	VolumeTotal          int
	InsertDate           string
	InsertTime           string
	ActiveTime           string
	SuspendTime          string
	UpdateTime           string
	CancelTime           string
	ActiveTraderID       string
	ClearingPartID       string
	SequenceNo           int
	FrontID              int
	SessionID            int
	UserProductInfo      string
	StatusMsg            string
	UserForceClose       bool
	ActiveUserID         string
	BrokerOrderSeq       int
	RelativeOrderSysID   string
	ZCETotalTradedVolume int
	IsSwapOrder          bool
	BranchID             string
	InvestUnitID         string
	AccountID            string
	CurrencyID           string
	IPAddress            string
	MACAddress           string
}

func (ord *Order) Identity() string {
	out := NewStringBuffer()
	defer out.Release()

	if ord.BrokerID != "" {
		out.WriteString(ord.BrokerID)
		out.WriteString(".")
	}

	out.WriteString(ord.InvestorID)
	out.WriteString(".")
	out.WriteString(strconv.Itoa(ord.SessionID))
	out.WriteString(".")
	out.WriteString(ord.OrderLocalID)

	return out.String()
}

type Trade struct {
	BrokerID       string
	InvestorID     string
	InstrumentID   string
	OrderRef       string
	UserID         string
	ExchangeID     string
	TradeID        string
	Direction      Direction
	OrderSysID     string
	ParticipantID  string
	ClientID       string
	TradingRole    TradingRole
	ExchangeInstID string
	OffsetFlag     OffsetFlag
	HedgeFlag      HedgeFlag
	Price          float64
	Volume         int
	TradeDate      string
	TradeTime      string
	TradeType      TradeType
	PriceSource    PriceSource
	TraderID       string
	OrderLocalID   string
	ClearingPartID string
	BusinessUnit   string
	SequenceNo     int
	TradingDay     string
	SettlementID   int
	BrokerOrderSeq int
	TradeSource    TradeSource
	InvestUnitID   string
}

type SubInfo struct {
	InvestorID  string
	AccountType AccountType
	BrokerID    string
	SubInfoType SubInfoType
}

type QryOrder struct {
	BrokerID        string
	InvestorID      string
	InstrumentID    string
	ExchangeID      string
	OrderSysID      string
	InsertTimeStart string
	InsertTimeEnd   string
	InvestUnitID    string
	FrontID         int
	SessionID       int
	OrderRef        string
}

type QryTrade struct {
	BrokerID        string
	InvestorID      string
	InstrumentID    string
	ExchangeID      string
	TradeID         string
	InsertTimeStart string
	InsertTimeEnd   string
	InvestUnitID    string
}

type Instrument struct {
	InstrumentID                 string
	ExchangeID                   string
	InstrumentName               string
	ExchangeInstID               string
	ProductID                    string
	ProductClass                 ProductClass
	DeliveryYear                 int
	DeliveryMonth                int
	MaxMarketOrderVolume         int
	MinMarketOrderVolume         int
	MaxLimitOrderVolume          int
	MinLimitOrderVolume          int
	VolumeMultiple               int
	PriceTick                    float64
	CreateDate                   string
	OpenDate                     string
	ExpireDate                   string
	StartDelivDate               string
	EndDelivDate                 string
	InstLiftPhase                InstLifePhase
	IsTrading                    bool
	PositionType                 PositionType
	PositionDateType             PositionDateType
	LongMarginRatioByMoney       float64
	LongMarginRatioByVolume      float64
	ShortMarginRatioByMoney      float64
	ShortMarginRatioByVolume     float64
	AfternoonMarginRatioByMoney  float64
	AfternoonMarginRatioByVolume float64
	OvernightMarginRatioByMoney  float64
	OvernightMarginRatioByVolume float64
	MarginIsRelative             bool
	OpenRatioByMoney             float64
	OpenRatioByVolume            float64
	CloseRatioByMoney            float64
	CloseRatioByVolume           float64
	CloseTodayRatioByMoney       float64
	CloseTodayRatioByVolume      float64
	UnderlyingInstID             string
	StrikePrice                  float64
	OptionsType                  OptionsType
	UnderlyingMultiple           float64
	CombinationType              CombinationType
}

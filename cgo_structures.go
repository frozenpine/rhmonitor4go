package rhmonitor4go

/*
#cgo CFLAGS: -I${SRCDIR}/include

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"bytes"
	"strconv"
	"sync"
	"unsafe"
)

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

var rspInfoCache = sync.Pool{New: func() any { return &RspInfo{} }}

func NewFromCRHRspInfoField(pRHRspInfoField *C.struct_CRHRspInfoField) *RspInfo {
	if pRHRspInfoField == nil {
		return nil
	}

	rsp := rspInfoCache.Get().(*RspInfo)

	rsp.ErrorID = int(pRHRspInfoField.ErrorID)
	rsp.ErrorMsg = CStr2GoStr(unsafe.Pointer(&pRHRspInfoField.ErrorMsg))

	return rsp
}

type RiskUser struct {
	UserID     string
	Password   string
	MACAddress string
}

func (usr RiskUser) IsValid() bool {
	return usr.UserID != "" && usr.Password != ""
}

func (usr RiskUser) ToCRHMonitorReqUserLoginField() *C.struct_CRHMonitorReqUserLoginField {
	data := C.struct_CRHMonitorReqUserLoginField{}

	C.memcpy(
		unsafe.Pointer(&data.UserID[0]),
		unsafe.Pointer(C.CString(usr.UserID)),
		C.sizeof_TRHUserIDType-1,
	)
	C.memcpy(
		unsafe.Pointer(&data.Password[0]),
		unsafe.Pointer(C.CString(usr.Password)),
		C.sizeof_TRHPasswordType-1,
	)

	return &data
}

func (usr RiskUser) ToCRHMonitorUserLogoutField() *C.struct_CRHMonitorUserLogoutField {
	data := C.struct_CRHMonitorUserLogoutField{}

	C.memcpy(
		unsafe.Pointer(&data.UserID),
		unsafe.Pointer(C.CString(usr.UserID)),
		C.sizeof_TRHUserIDType-1,
	)

	return &data
}

func (usr RiskUser) ToCRHMonitorQryMonitorUser() *C.struct_CRHMonitorQryMonitorUser {
	data := C.struct_CRHMonitorQryMonitorUser{}

	C.memcpy(
		unsafe.Pointer(&data.UserID),
		unsafe.Pointer(C.CString(usr.UserID)),
		C.sizeof_TRHUserIDType-1,
	)

	return &data
}

type RspUserLogin struct {
	UserID            string
	PrivilegeType     PrivilegeType
	InfoPrivilegeType string
	TradingDay        string
	LoginTime         string
}

func NewFromCRHMonitorRspUserLoginField(pRspUserLoginField *C.struct_CRHMonitorRspUserLoginField) *RspUserLogin {
	if pRspUserLoginField == nil {
		return nil
	}

	return &RspUserLogin{
		UserID:            CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.UserID)),
		PrivilegeType:     PrivilegeType(pRspUserLoginField.PrivilegeType),
		InfoPrivilegeType: CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.InfoPrivilegeType)),
		TradingDay:        CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.TradingDay)),
		LoginTime:         CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.LoginTime)),
	}
}

type RspUserLogout struct {
	UserID string
}

func NewFromCRHMonitorUserLogoutField(pRspUserLoginField *C.struct_CRHMonitorUserLogoutField) *RspUserLogout {
	if pRspUserLoginField == nil {
		return nil
	}

	return &RspUserLogout{
		UserID: CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.UserID)),
	}
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

func (i Investor) ToCRHMonitorQryInvestorMoneyField() *C.struct_CRHMonitorQryInvestorMoneyField {
	data := C.struct_CRHMonitorQryInvestorMoneyField{}

	C.memcpy(
		unsafe.Pointer(&data.InvestorID),
		unsafe.Pointer(C.CString(i.InvestorID)),
		C.sizeof_TRHInvestorIDType-1,
	)

	C.memcpy(
		unsafe.Pointer(&data.BrokerID),
		unsafe.Pointer(C.CString(i.BrokerID)),
		C.sizeof_TRHBrokerIDType-1,
	)

	return &data
}

func (i Investor) ToCRHMonitorQryInvestorPositionField(instrumentID string) *C.struct_CRHMonitorQryInvestorPositionField {
	data := C.struct_CRHMonitorQryInvestorPositionField{}

	C.memcpy(
		unsafe.Pointer(&data.InvestorID),
		unsafe.Pointer(C.CString(i.InvestorID)),
		C.sizeof_TRHInvestorIDType-1,
	)

	C.memcpy(
		unsafe.Pointer(&data.BrokerID),
		unsafe.Pointer(C.CString(i.BrokerID)),
		C.sizeof_TRHBrokerIDType-1,
	)

	if instrumentID != "" {
		C.memcpy(
			unsafe.Pointer(&data.InstrumentID),
			unsafe.Pointer(C.CString(instrumentID)),
			C.sizeof_TRHInstrumentIDType-1,
		)
	}

	return &data
}

var investorCache = sync.Pool{New: func() any { return &Investor{} }}

func NewFromCRHQryInvestorField(pRspMonitorUser *C.struct_CRHQryInvestorField) *Investor {
	if pRspMonitorUser == nil {
		return nil
	}

	investor := investorCache.Get().(*Investor)

	investor.BrokerID = CStr2GoStr(unsafe.Pointer(&pRspMonitorUser.BrokerID))
	investor.InvestorID = CStr2GoStr(unsafe.Pointer(&pRspMonitorUser.InvestorID))

	return investor
}

type Account struct {
	//??????????????????
	BrokerID string
	//???????????????
	AccountID string
	//??????????????????
	PreMortgage float64
	//??????????????????
	PreCredit float64
	//???????????????
	PreDeposit float64
	//?????????????????????
	PreBalance float64
	//????????????????????????
	PreMargin float64
	//????????????
	InterestBase float64
	//????????????
	Interest float64
	//????????????
	Deposit float64
	//????????????
	Withdraw float64
	//??????????????????
	FrozenMargin float64
	//???????????????
	FrozenCash float64
	//??????????????????
	FrozenCommission float64
	//?????????????????????
	CurrMargin float64
	//????????????
	CashIn float64
	//?????????
	Commission float64
	//????????????
	CloseProfit float64
	//????????????
	PositionProfit float64
	//?????????????????????
	Balance float64
	//????????????
	Available float64
	//????????????
	WithdrawQuota float64
	//???????????????
	Reserve float64
	//?????????
	TradingDay string
	//????????????
	SettlementID int
	//????????????
	Credit float64
	//????????????
	Mortgage float64
	//??????????????????
	ExchangeMargin float64
	//????????????????????????
	DeliveryMargin float64
	//????????????????????????
	ExchangeDeliveryMargin float64
	//???????????????????????????
	ReserveBalance float64
	//????????????
	CurrencyID string
	//????????????????????????
	PreFundMortgageIn float64
	//????????????????????????
	PreFundMortgageOut float64
	//??????????????????
	FundMortgageIn float64
	//??????????????????
	FundMortgageOut float64
	//??????????????????
	FundMortgageAvailable float64
	//?????????????????????
	MortgageableFund float64
	//???????????????????????????
	SpecProductMargin float64
	//???????????????????????????
	SpecProductFrozenMargin float64
	//?????????????????????
	SpecProductCommission float64
	//???????????????????????????
	SpecProductFrozenCommission float64
	//????????????????????????
	SpecProductPositionProfit float64
	//????????????????????????
	SpecProductCloseProfit float64
	//?????????????????????????????????????????????????????????
	SpecProductPositionProfitByAlg float64
	//??????????????????????????????
	SpecProductExchangeMargin float64
	//????????????
	BizType BusinessType
	//????????????????????????
	FrozenSwap float64
	//??????????????????
	RemainSwap float64
	//??????????????????
	TotalStockMarketValue float64
	//??????????????????
	TotalOptionMarketValue float64
	//????????????
	DynamicMoney float64
	//???????????????
	Premium float64
	//????????????
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

var accountCache = sync.Pool{New: func() any { return &Account{} }}

func NewFromCRHTradingAccountField(pRHTradingAccountField *C.struct_CRHTradingAccountField) *Account {
	if pRHTradingAccountField == nil {
		return nil
	}

	account := accountCache.Get().(*Account)

	account.BrokerID = CStr2GoStr(unsafe.Pointer(&pRHTradingAccountField.BrokerID))
	account.AccountID = CStr2GoStr(unsafe.Pointer(&pRHTradingAccountField.AccountID))
	account.PreMortgage = float64(pRHTradingAccountField.PreMortgage)
	account.PreCredit = float64(pRHTradingAccountField.PreCredit)
	account.PreDeposit = float64(pRHTradingAccountField.PreDeposit)
	account.PreBalance = float64(pRHTradingAccountField.PreBalance)
	account.PreMargin = float64(pRHTradingAccountField.PreMargin)
	account.InterestBase = float64(pRHTradingAccountField.InterestBase)
	account.Interest = float64(pRHTradingAccountField.Interest)
	account.Deposit = float64(pRHTradingAccountField.Deposit)
	account.Withdraw = float64(pRHTradingAccountField.Withdraw)
	account.FrozenMargin = float64(pRHTradingAccountField.FrozenMargin)
	account.FrozenCash = float64(pRHTradingAccountField.FrozenCash)
	account.FrozenCommission = float64(pRHTradingAccountField.FrozenCommission)
	account.CurrMargin = float64(pRHTradingAccountField.CurrMargin)
	account.CashIn = float64(pRHTradingAccountField.CashIn)
	account.Commission = float64(pRHTradingAccountField.Commission)
	account.CloseProfit = float64(pRHTradingAccountField.CloseProfit)
	account.PositionProfit = float64(pRHTradingAccountField.PositionProfit)
	account.Balance = float64(pRHTradingAccountField.Balance)
	account.Available = float64(pRHTradingAccountField.Available)
	account.WithdrawQuota = float64(pRHTradingAccountField.WithdrawQuota)
	account.Reserve = float64(pRHTradingAccountField.Reserve)
	account.TradingDay = CStr2GoStr(unsafe.Pointer(&pRHTradingAccountField.TradingDay))
	account.SettlementID = int(pRHTradingAccountField.SettlementID)
	account.Credit = float64(pRHTradingAccountField.Credit)
	account.Mortgage = float64(pRHTradingAccountField.Mortgage)
	account.ExchangeMargin = float64(pRHTradingAccountField.ExchangeMargin)
	account.DeliveryMargin = float64(pRHTradingAccountField.DeliveryMargin)
	account.ExchangeDeliveryMargin = float64(pRHTradingAccountField.ExchangeDeliveryMargin)
	account.ReserveBalance = float64(pRHTradingAccountField.ReserveBalance)
	account.CurrencyID = CStr2GoStr(unsafe.Pointer(&pRHTradingAccountField.CurrencyID))
	account.PreFundMortgageIn = float64(pRHTradingAccountField.PreFundMortgageIn)
	account.PreFundMortgageOut = float64(pRHTradingAccountField.PreFundMortgageOut)
	account.FundMortgageIn = float64(pRHTradingAccountField.FundMortgageIn)
	account.FundMortgageOut = float64(pRHTradingAccountField.FundMortgageOut)
	account.FundMortgageAvailable = float64(pRHTradingAccountField.FundMortgageAvailable)
	account.MortgageableFund = float64(pRHTradingAccountField.MortgageableFund)
	account.SpecProductMargin = float64(pRHTradingAccountField.SpecProductMargin)
	account.SpecProductFrozenMargin = float64(pRHTradingAccountField.SpecProductFrozenMargin)
	account.SpecProductCommission = float64(pRHTradingAccountField.SpecProductCommission)
	account.SpecProductFrozenCommission = float64(pRHTradingAccountField.SpecProductFrozenCommission)
	account.SpecProductPositionProfit = float64(pRHTradingAccountField.SpecProductPositionProfit)
	account.SpecProductCloseProfit = float64(pRHTradingAccountField.SpecProductCloseProfit)
	account.SpecProductPositionProfitByAlg = float64(pRHTradingAccountField.SpecProductPositionProfitByAlg)
	account.SpecProductExchangeMargin = float64(pRHTradingAccountField.SpecProductExchangeMargin)
	account.BizType = BusinessType(pRHTradingAccountField.BizType)
	account.FrozenSwap = float64(pRHTradingAccountField.FrozenSwap)
	account.RemainSwap = float64(pRHTradingAccountField.RemainSwap)
	account.TotalStockMarketValue = float64(pRHTradingAccountField.TotalStockMarketValue)
	account.TotalOptionMarketValue = float64(pRHTradingAccountField.TotalOptionMarketValue)
	account.DynamicMoney = float64(pRHTradingAccountField.DynamicMoney)
	account.Premium = float64(pRHTradingAccountField.Premium)
	account.MarketValueEquity = float64(pRHTradingAccountField.MarketValueEquity)

	return account
}

type Position struct {
	//???????????????
	InvestorID string
	//??????????????????
	BrokerID string
	//????????????
	ProductID string
	//????????????
	InstrumentID string
	//??????????????????
	HedgeFlag HedgeFlag
	//????????????
	Direction Direction
	//????????????
	Volume int
	//???????????????
	Margin float64
	//??????????????????
	AvgOpenPriceByVol float64
	//??????????????????
	AvgOpenPrice float64
	//????????????
	TodayVolume int
	//??????????????????
	FrozenVolume int
	//????????????
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

var positionCache = sync.Pool{New: func() any { return &Position{} }}

func NewFromCRHMonitorPositionField(pRHMonitorPositionField *C.struct_CRHMonitorPositionField) *Position {
	if pRHMonitorPositionField == nil {
		return nil
	}

	pos := positionCache.Get().(*Position)

	pos.InvestorID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.InvestorID))
	pos.BrokerID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.BrokerID))
	pos.ProductID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.ProductID))
	pos.InstrumentID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.InstrumentID))
	pos.HedgeFlag = HedgeFlag(pRHMonitorPositionField.HedgeFlag)
	pos.Direction = Direction(pRHMonitorPositionField.Direction)
	pos.Volume = int(pRHMonitorPositionField.Volume)
	pos.Margin = float64(pRHMonitorPositionField.Margin)
	pos.AvgOpenPriceByVol = float64(pRHMonitorPositionField.AvgOpenPriceByVol)
	pos.AvgOpenPrice = float64(pRHMonitorPositionField.AvgOpenPrice)
	pos.TodayVolume = int(pRHMonitorPositionField.TodayVolume)
	pos.FrozenVolume = int(pRHMonitorPositionField.FrozenVolume)
	pos.EntryType = uint8(pRHMonitorPositionField.EntryType)

	return pos
}

type OffsetOrder struct {
	//?????????
	InvestorID string
	//??????????????????
	BrokerID string
	//??????ID
	InstrumentID string
	//??????
	Direction Direction
	//??????
	Volume int
	//??????
	Price float64
	//??????????????????
	ComboOffsetFlag [5]byte
	//????????????????????????
	ComboHedgeFlag [5]byte
}

var offsetOrderCache = sync.Pool{New: func() any { return &OffsetOrder{} }}

func NewFromCRHMonitorOffsetOrderField(pMonitorOrderField *C.struct_CRHMonitorOffsetOrderField) *OffsetOrder {
	if pMonitorOrderField == nil {
		return nil
	}

	offsetOrd := offsetOrderCache.Get().(*OffsetOrder)

	offsetOrd.InvestorID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.InvestorID))
	offsetOrd.BrokerID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.BrokerID))
	offsetOrd.InstrumentID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.InstrumentID))
	offsetOrd.Direction = Direction(pMonitorOrderField.Direction)
	offsetOrd.Volume = int(pMonitorOrderField.volume)
	offsetOrd.Price = float64(pMonitorOrderField.Price)

	CopyN(offsetOrd.ComboOffsetFlag[:], unsafe.Pointer(&pMonitorOrderField.CombOffsetFlag), 5)

	CopyN(offsetOrd.ComboHedgeFlag[:], unsafe.Pointer(&pMonitorOrderField.CombHedgeFlag), 5)

	return offsetOrd
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

var orderCache = sync.Pool{New: func() any { return &Order{} }}

func NewFromCRHOrderField(pOrder *C.struct_CRHOrderField) *Order {
	if pOrder == nil {
		return nil
	}

	ord := orderCache.Get().(*Order)

	ord.BrokerID = CStr2GoStr(unsafe.Pointer(&pOrder.BrokerID))
	ord.InvestorID = CStr2GoStr(unsafe.Pointer(&pOrder.InvestorID))
	ord.InstrumentID = CStr2GoStr(unsafe.Pointer(&pOrder.InstrumentID))
	ord.OrderRef = CStr2GoStr(unsafe.Pointer(&pOrder.OrderRef))
	ord.UserID = CStr2GoStr(unsafe.Pointer(&pOrder.UserID))
	ord.PriceType = OrderPriceType(pOrder.OrderPriceType)
	ord.Direction = Direction(pOrder.Direction)
	CopyN(ord.ComboOffsetFlag[:], unsafe.Pointer(&pOrder.CombOffsetFlag), 5)
	CopyN(ord.ComboHedgeFlag[:], unsafe.Pointer(&pOrder.CombHedgeFlag), 5)
	ord.LimitPrice = float64(pOrder.LimitPrice)
	ord.VolumeTotalOriginal = int(pOrder.VolumeTotalOriginal)
	ord.TimeCondition = TimeCondition(pOrder.TimeCondition)
	ord.GTDDate = CStr2GoStr(unsafe.Pointer(&pOrder.GTDDate))
	ord.VolumeCondition = VolumeCondition(pOrder.VolumeCondition)
	ord.MinVolume = int(pOrder.MinVolume)
	ord.ContingentCondition = ContingentCondition(pOrder.ContingentCondition)
	ord.StopPrice = float64(pOrder.StopPrice)
	ord.ForceCloseReason = ForceCloseReason(pOrder.ForceCloseReason)
	ord.IsAutoSuspend = pOrder.IsAutoSuspend != 0
	ord.BusinessUnit = CStr2GoStr(unsafe.Pointer(&pOrder.BusinessUnit))
	ord.RequestID = int(pOrder.RequestID)
	ord.OrderLocalID = CStr2GoStr(unsafe.Pointer(&pOrder.OrderLocalID))
	ord.ExchangeID = CStr2GoStr(unsafe.Pointer(&pOrder.ExchangeID))
	ord.ParticipantID = CStr2GoStr(unsafe.Pointer(&pOrder.ParticipantID))
	ord.ClientID = CStr2GoStr(unsafe.Pointer(&pOrder.ClientID))
	ord.ExchangeInstID = CStr2GoStr(unsafe.Pointer(&pOrder.ExchangeInstID))
	ord.TraderID = CStr2GoStr(unsafe.Pointer(&pOrder.TraderID))
	ord.InstallID = int(pOrder.InstallID)
	ord.OrderSubmitStatus = OrderSubmitStatus(pOrder.OrderSubmitStatus)
	ord.NotifySequence = int(pOrder.NotifySequence)
	ord.TradingDay = CStr2GoStr(unsafe.Pointer(&pOrder.TradingDay))
	ord.SettlementID = int(pOrder.SettlementID)
	ord.OrderSysID = CStr2GoStr(unsafe.Pointer(&pOrder.OrderSysID))
	ord.OrderSource = OrderSource(pOrder.OrderSource)
	ord.OrderStatus = OrderStatus(pOrder.OrderStatus)
	ord.OrderType = OrderType(pOrder.OrderType)
	ord.VolumeTraded = int(pOrder.VolumeTraded)
	ord.VolumeTotal = int(pOrder.VolumeTotal)
	ord.InsertDate = CStr2GoStr(unsafe.Pointer(&pOrder.InsertDate))
	ord.InsertTime = CStr2GoStr(unsafe.Pointer(&pOrder.InsertTime))
	ord.ActiveTime = CStr2GoStr(unsafe.Pointer(&pOrder.ActiveTime))
	ord.SuspendTime = CStr2GoStr(unsafe.Pointer(&pOrder.SuspendTime))
	ord.UpdateTime = CStr2GoStr(unsafe.Pointer(&pOrder.UpdateTime))
	ord.CancelTime = CStr2GoStr(unsafe.Pointer(&pOrder.CancelTime))
	ord.ActiveTraderID = CStr2GoStr(unsafe.Pointer(&pOrder.ActiveTraderID))
	ord.ClearingPartID = CStr2GoStr(unsafe.Pointer(&pOrder.ClearingPartID))
	ord.SequenceNo = int(pOrder.SequenceNo)
	ord.FrontID = int(pOrder.FrontID)
	ord.SessionID = int(pOrder.SessionID)
	ord.UserProductInfo = CStr2GoStr(unsafe.Pointer(&pOrder.UserProductInfo))
	ord.StatusMsg = CStr2GoStr(unsafe.Pointer(&pOrder.StatusMsg))
	ord.UserForceClose = pOrder.UserForceClose != 0
	ord.ActiveUserID = CStr2GoStr(unsafe.Pointer(&pOrder.ActiveUserID))
	ord.BrokerOrderSeq = int(pOrder.BrokerOrderSeq)
	ord.RelativeOrderSysID = CStr2GoStr(unsafe.Pointer(&pOrder.RelativeOrderSysID))
	ord.ZCETotalTradedVolume = int(pOrder.ZCETotalTradedVolume)
	ord.IsSwapOrder = pOrder.IsSwapOrder != 0
	ord.BranchID = CStr2GoStr(unsafe.Pointer(&pOrder.BranchID))
	ord.InvestUnitID = CStr2GoStr(unsafe.Pointer(&pOrder.InvestUnitID))
	ord.AccountID = CStr2GoStr(unsafe.Pointer(&pOrder.AccountID))
	ord.CurrencyID = CStr2GoStr(unsafe.Pointer(&pOrder.CurrencyID))
	ord.IPAddress = CStr2GoStr(unsafe.Pointer(&pOrder.IPAddress))
	ord.MACAddress = CStr2GoStr(unsafe.Pointer(&pOrder.MacAddress))

	return ord
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

var tradeCache = sync.Pool{New: func() any { return &Trade{} }}

func NewFromCRHTradeField(pTrade *C.struct_CRHTradeField) *Trade {
	if pTrade == nil {
		return nil
	}

	td := tradeCache.Get().(*Trade)

	td.BrokerID = CStr2GoStr(unsafe.Pointer(&pTrade.BrokerID))
	td.InvestorID = CStr2GoStr(unsafe.Pointer(&pTrade.InvestorID))
	td.InstrumentID = CStr2GoStr(unsafe.Pointer(&pTrade.InstrumentID))
	td.OrderRef = CStr2GoStr(unsafe.Pointer(&pTrade.OrderRef))
	td.UserID = CStr2GoStr(unsafe.Pointer(&pTrade.UserID))
	td.ExchangeID = CStr2GoStr(unsafe.Pointer(&pTrade.ExchangeID))
	td.TradeID = CStr2GoStr(unsafe.Pointer(&pTrade.TradeID))
	td.Direction = Direction(pTrade.Direction)
	td.OrderSysID = CStr2GoStr(unsafe.Pointer(&pTrade.OrderSysID))
	td.ParticipantID = CStr2GoStr(unsafe.Pointer(&pTrade.ParticipantID))
	td.ClientID = CStr2GoStr(unsafe.Pointer(&pTrade.ClientID))
	td.TradingRole = TradingRole(pTrade.TradingRole)
	td.ExchangeInstID = CStr2GoStr(unsafe.Pointer(&pTrade.ExchangeInstID))
	td.OffsetFlag = OffsetFlag(pTrade.OffsetFlag)
	td.HedgeFlag = HedgeFlag(pTrade.HedgeFlag)
	td.Price = float64(pTrade.Price)
	td.Volume = int(pTrade.Volume)
	td.TradeDate = CStr2GoStr(unsafe.Pointer(&pTrade.TradeDate))
	td.TradeTime = CStr2GoStr(unsafe.Pointer(&pTrade.TradeTime))
	td.TradeType = TradeType(pTrade.TradeType)
	td.PriceSource = PriceSource(pTrade.PriceSource)
	td.TraderID = CStr2GoStr(unsafe.Pointer(&pTrade.TraderID))
	td.OrderLocalID = CStr2GoStr(unsafe.Pointer(&pTrade.OrderLocalID))
	td.ClearingPartID = CStr2GoStr(unsafe.Pointer(&pTrade.ClearingPartID))
	td.BusinessUnit = CStr2GoStr(unsafe.Pointer(&pTrade.BusinessUnit))
	td.SequenceNo = int(pTrade.SequenceNo)
	td.TradingDay = CStr2GoStr(unsafe.Pointer(&pTrade.TradingDay))
	td.SettlementID = int(pTrade.SettlementID)
	td.BrokerOrderSeq = int(pTrade.BrokerOrderSeq)
	td.TradeSource = TradeSource(pTrade.TradeSource)
	td.InvestUnitID = CStr2GoStr(unsafe.Pointer(&pTrade.InvestUnitID))

	return td
}

type SubInfo struct {
	InvestorID  string
	AccountType AccountType
	BrokerID    string
	SubInfoType SubInfoType
}

func (sub SubInfo) ToCRHMonitorSubPushInfo() *C.struct_CRHMonitorSubPushInfo {
	data := C.struct_CRHMonitorSubPushInfo{}

	C.memcpy(
		unsafe.Pointer(&data.InvestorID),
		unsafe.Pointer(C.CString(sub.InvestorID)),
		C.sizeof_TRHInvestorIDType-1,
	)

	data.AccountType = (C.TRHAccountType)(sub.AccountType)

	C.memcpy(
		unsafe.Pointer(&data.BrokerID),
		unsafe.Pointer(C.CString(sub.BrokerID)),
		C.sizeof_TRHBrokerIDType-1,
	)

	data.SubInfoType = (C.RHMonitorSubPushInfoType)(sub.SubInfoType)

	return &data
}

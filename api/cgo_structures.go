package api

/*
#cgo CFLAGS: -I${SRCDIR}/include

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"sync"
	"unsafe"

	"github.com/frozenpine/rhmonitor4go"
)

var rspInfoCache = sync.Pool{New: func() any { return &rhmonitor4go.RspInfo{} }}

func NewFromCRHRspInfoField(pRHRspInfoField *C.struct_CRHRspInfoField) *rhmonitor4go.RspInfo {
	if pRHRspInfoField == nil {
		return nil
	}

	rsp := rspInfoCache.Get().(*rhmonitor4go.RspInfo)

	rsp.ErrorID = int(pRHRspInfoField.ErrorID)
	rsp.ErrorMsg = CStr2GoStr(unsafe.Pointer(&pRHRspInfoField.ErrorMsg))

	return rsp
}

func ToCRHMonitorReqUserLoginField(usr *rhmonitor4go.RiskUser) *C.struct_CRHMonitorReqUserLoginField {
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

func ToCRHMonitorUserLogoutField(usr *rhmonitor4go.RiskUser) *C.struct_CRHMonitorUserLogoutField {
	data := C.struct_CRHMonitorUserLogoutField{}

	C.memcpy(
		unsafe.Pointer(&data.UserID),
		unsafe.Pointer(C.CString(usr.UserID)),
		C.sizeof_TRHUserIDType-1,
	)

	return &data
}

func ToCRHMonitorQryMonitorUser(usr *rhmonitor4go.RiskUser) *C.struct_CRHMonitorQryMonitorUser {
	data := C.struct_CRHMonitorQryMonitorUser{}

	C.memcpy(
		unsafe.Pointer(&data.UserID),
		unsafe.Pointer(C.CString(usr.UserID)),
		C.sizeof_TRHUserIDType-1,
	)

	return &data
}

func NewFromCRHMonitorRspUserLoginField(pRspUserLoginField *C.struct_CRHMonitorRspUserLoginField) *rhmonitor4go.RspUserLogin {
	if pRspUserLoginField == nil {
		return nil
	}

	login := rhmonitor4go.RspUserLogin{
		UserID:        CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.UserID)),
		PrivilegeType: rhmonitor4go.PrivilegeType(pRspUserLoginField.PrivilegeType),
		TradingDay:    CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.TradingDay)),
		LoginTime:     CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.LoginTime)),
	}

	CopyN(login.InfoPrivilegeType[:], unsafe.Pointer(&pRspUserLoginField.InfoPrivilegeType), 200)

	return &login
}

func NewFromCRHMonitorUserLogoutField(pRspUserLoginField *C.struct_CRHMonitorUserLogoutField) *rhmonitor4go.RspUserLogout {
	if pRspUserLoginField == nil {
		return nil
	}

	return &rhmonitor4go.RspUserLogout{
		UserID: CStr2GoStr(unsafe.Pointer(&pRspUserLoginField.UserID)),
	}
}

func ToCRHMonitorQryInvestorMoneyField(i *rhmonitor4go.Investor) *C.struct_CRHMonitorQryInvestorMoneyField {
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

func ToCRHMonitorQryInvestorPositionField(i *rhmonitor4go.Investor, instrumentID string) *C.struct_CRHMonitorQryInvestorPositionField {
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

var investorCache = sync.Pool{New: func() any { return &rhmonitor4go.Investor{} }}

func NewFromCRHQryInvestorField(pRspMonitorUser *C.struct_CRHQryInvestorField) *rhmonitor4go.Investor {
	if pRspMonitorUser == nil {
		return nil
	}

	investor := investorCache.Get().(*rhmonitor4go.Investor)

	investor.BrokerID = CStr2GoStr(unsafe.Pointer(&pRspMonitorUser.BrokerID))
	investor.InvestorID = CStr2GoStr(unsafe.Pointer(&pRspMonitorUser.InvestorID))

	return investor
}

var accountCache = sync.Pool{New: func() any { return &rhmonitor4go.Account{} }}

func NewFromCRHTradingAccountField(pRHTradingAccountField *C.struct_CRHTradingAccountField) *rhmonitor4go.Account {
	if pRHTradingAccountField == nil {
		return nil
	}

	account := accountCache.Get().(*rhmonitor4go.Account)

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
	account.BizType = rhmonitor4go.BusinessType(pRHTradingAccountField.BizType)
	account.FrozenSwap = float64(pRHTradingAccountField.FrozenSwap)
	account.RemainSwap = float64(pRHTradingAccountField.RemainSwap)
	account.TotalStockMarketValue = float64(pRHTradingAccountField.TotalStockMarketValue)
	account.TotalOptionMarketValue = float64(pRHTradingAccountField.TotalOptionMarketValue)
	account.DynamicMoney = float64(pRHTradingAccountField.DynamicMoney)
	account.Premium = float64(pRHTradingAccountField.Premium)
	account.MarketValueEquity = float64(pRHTradingAccountField.MarketValueEquity)

	return account
}

var positionCache = sync.Pool{New: func() any { return &rhmonitor4go.Position{} }}

func NewFromCRHMonitorPositionField(pRHMonitorPositionField *C.struct_CRHMonitorPositionField) *rhmonitor4go.Position {
	if pRHMonitorPositionField == nil {
		return nil
	}

	pos := positionCache.Get().(*rhmonitor4go.Position)

	pos.InvestorID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.InvestorID))
	pos.BrokerID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.BrokerID))
	pos.ProductID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.ProductID))
	pos.InstrumentID = CStr2GoStr(unsafe.Pointer(&pRHMonitorPositionField.InstrumentID))
	pos.HedgeFlag = rhmonitor4go.HedgeFlag(pRHMonitorPositionField.HedgeFlag)
	pos.Direction = rhmonitor4go.Direction(pRHMonitorPositionField.Direction)
	pos.Volume = int(pRHMonitorPositionField.Volume)
	pos.Margin = float64(pRHMonitorPositionField.Margin)
	pos.AvgOpenPriceByVol = float64(pRHMonitorPositionField.AvgOpenPriceByVol)
	pos.AvgOpenPrice = float64(pRHMonitorPositionField.AvgOpenPrice)
	pos.TodayVolume = int(pRHMonitorPositionField.TodayVolume)
	pos.FrozenVolume = int(pRHMonitorPositionField.FrozenVolume)
	pos.EntryType = uint8(pRHMonitorPositionField.EntryType)

	return pos
}

var offsetOrderCache = sync.Pool{New: func() any { return &rhmonitor4go.OffsetOrder{} }}

func NewFromCRHMonitorOffsetOrderField(pMonitorOrderField *C.struct_CRHMonitorOffsetOrderField) *rhmonitor4go.OffsetOrder {
	if pMonitorOrderField == nil {
		return nil
	}

	offsetOrd := offsetOrderCache.Get().(*rhmonitor4go.OffsetOrder)

	offsetOrd.InvestorID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.InvestorID))
	offsetOrd.BrokerID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.BrokerID))
	offsetOrd.InstrumentID = CStr2GoStr(unsafe.Pointer(&pMonitorOrderField.InstrumentID))
	offsetOrd.Direction = rhmonitor4go.Direction(pMonitorOrderField.Direction)
	offsetOrd.Volume = int(pMonitorOrderField.volume)
	offsetOrd.Price = float64(pMonitorOrderField.Price)

	CopyN(offsetOrd.ComboOffsetFlag[:], unsafe.Pointer(&pMonitorOrderField.CombOffsetFlag), 5)

	CopyN(offsetOrd.ComboHedgeFlag[:], unsafe.Pointer(&pMonitorOrderField.CombHedgeFlag), 5)

	return offsetOrd
}

var orderCache = sync.Pool{New: func() any { return &rhmonitor4go.Order{} }}

func NewFromCRHOrderField(pOrder *C.struct_CRHOrderField) *rhmonitor4go.Order {
	if pOrder == nil {
		return nil
	}

	ord := orderCache.Get().(*rhmonitor4go.Order)

	ord.BrokerID = CStr2GoStr(unsafe.Pointer(&pOrder.BrokerID))
	ord.InvestorID = CStr2GoStr(unsafe.Pointer(&pOrder.InvestorID))
	ord.InstrumentID = CStr2GoStr(unsafe.Pointer(&pOrder.InstrumentID))
	ord.OrderRef = CStr2GoStr(unsafe.Pointer(&pOrder.OrderRef))
	ord.UserID = CStr2GoStr(unsafe.Pointer(&pOrder.UserID))
	ord.PriceType = rhmonitor4go.OrderPriceType(pOrder.OrderPriceType)
	ord.Direction = rhmonitor4go.Direction(pOrder.Direction)
	CopyN(ord.ComboOffsetFlag[:], unsafe.Pointer(&pOrder.CombOffsetFlag), 5)
	CopyN(ord.ComboHedgeFlag[:], unsafe.Pointer(&pOrder.CombHedgeFlag), 5)
	ord.LimitPrice = float64(pOrder.LimitPrice)
	ord.VolumeTotalOriginal = int(pOrder.VolumeTotalOriginal)
	ord.TimeCondition = rhmonitor4go.TimeCondition(pOrder.TimeCondition)
	ord.GTDDate = CStr2GoStr(unsafe.Pointer(&pOrder.GTDDate))
	ord.VolumeCondition = rhmonitor4go.VolumeCondition(pOrder.VolumeCondition)
	ord.MinVolume = int(pOrder.MinVolume)
	ord.ContingentCondition = rhmonitor4go.ContingentCondition(pOrder.ContingentCondition)
	ord.StopPrice = float64(pOrder.StopPrice)
	ord.ForceCloseReason = rhmonitor4go.ForceCloseReason(pOrder.ForceCloseReason)
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
	ord.OrderSubmitStatus = rhmonitor4go.OrderSubmitStatus(pOrder.OrderSubmitStatus)
	ord.NotifySequence = int(pOrder.NotifySequence)
	ord.TradingDay = CStr2GoStr(unsafe.Pointer(&pOrder.TradingDay))
	ord.SettlementID = int(pOrder.SettlementID)
	ord.OrderSysID = CStr2GoStr(unsafe.Pointer(&pOrder.OrderSysID))
	ord.OrderSource = rhmonitor4go.OrderSource(pOrder.OrderSource)
	ord.OrderStatus = rhmonitor4go.OrderStatus(pOrder.OrderStatus)
	ord.OrderType = rhmonitor4go.OrderType(pOrder.OrderType)
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

var tradeCache = sync.Pool{New: func() any { return &rhmonitor4go.Trade{} }}

func NewFromCRHTradeField(pTrade *C.struct_CRHTradeField) *rhmonitor4go.Trade {
	if pTrade == nil {
		return nil
	}

	td := tradeCache.Get().(*rhmonitor4go.Trade)

	td.BrokerID = CStr2GoStr(unsafe.Pointer(&pTrade.BrokerID))
	td.InvestorID = CStr2GoStr(unsafe.Pointer(&pTrade.InvestorID))
	td.InstrumentID = CStr2GoStr(unsafe.Pointer(&pTrade.InstrumentID))
	td.OrderRef = CStr2GoStr(unsafe.Pointer(&pTrade.OrderRef))
	td.UserID = CStr2GoStr(unsafe.Pointer(&pTrade.UserID))
	td.ExchangeID = CStr2GoStr(unsafe.Pointer(&pTrade.ExchangeID))
	td.TradeID = CStr2GoStr(unsafe.Pointer(&pTrade.TradeID))
	td.Direction = rhmonitor4go.Direction(pTrade.Direction)
	td.OrderSysID = CStr2GoStr(unsafe.Pointer(&pTrade.OrderSysID))
	td.ParticipantID = CStr2GoStr(unsafe.Pointer(&pTrade.ParticipantID))
	td.ClientID = CStr2GoStr(unsafe.Pointer(&pTrade.ClientID))
	td.TradingRole = rhmonitor4go.TradingRole(pTrade.TradingRole)
	td.ExchangeInstID = CStr2GoStr(unsafe.Pointer(&pTrade.ExchangeInstID))
	td.OffsetFlag = rhmonitor4go.OffsetFlag(pTrade.OffsetFlag)
	td.HedgeFlag = rhmonitor4go.HedgeFlag(pTrade.HedgeFlag)
	td.Price = float64(pTrade.Price)
	td.Volume = int(pTrade.Volume)
	td.TradeDate = CStr2GoStr(unsafe.Pointer(&pTrade.TradeDate))
	td.TradeTime = CStr2GoStr(unsafe.Pointer(&pTrade.TradeTime))
	td.TradeType = rhmonitor4go.TradeType(pTrade.TradeType)
	td.PriceSource = rhmonitor4go.PriceSource(pTrade.PriceSource)
	td.TraderID = CStr2GoStr(unsafe.Pointer(&pTrade.TraderID))
	td.OrderLocalID = CStr2GoStr(unsafe.Pointer(&pTrade.OrderLocalID))
	td.ClearingPartID = CStr2GoStr(unsafe.Pointer(&pTrade.ClearingPartID))
	td.BusinessUnit = CStr2GoStr(unsafe.Pointer(&pTrade.BusinessUnit))
	td.SequenceNo = int(pTrade.SequenceNo)
	td.TradingDay = CStr2GoStr(unsafe.Pointer(&pTrade.TradingDay))
	td.SettlementID = int(pTrade.SettlementID)
	td.BrokerOrderSeq = int(pTrade.BrokerOrderSeq)
	td.TradeSource = rhmonitor4go.TradeSource(pTrade.TradeSource)
	td.InvestUnitID = CStr2GoStr(unsafe.Pointer(&pTrade.InvestUnitID))

	return td
}

func ToCRHMonitorSubPushInfo(sub *rhmonitor4go.SubInfo) *C.struct_CRHMonitorSubPushInfo {
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

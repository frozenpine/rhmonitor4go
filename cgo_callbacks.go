package rhmonitor4go

/*
#cgo CFLAGS: -I${SRCDIR}/include

#include "cRHMonitorApi.h"

void cgoOnFrontConnected(CRHMonitorInstance instance);

void cgoOnFrontDisconnected(CRHMonitorInstance instance, int nReason);

void cgoOnRspUserLogin
(
	CRHMonitorInstance instance,
	struct CRHMonitorRspUserLoginField *pRspUserLoginField,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID
);

void cgoOnRspUserLogout
(
	CRHMonitorInstance instance,
	struct CRHMonitorUserLogoutField *pRspUserLoginField,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID
);

void cgoOnRspQryMonitorAccounts
(
	CRHMonitorInstance instance,
	struct CRHQryInvestorField *pRspMonitorUser,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID, bool isLast
);

void cgoOnRspQryInvestorMoney
(
	CRHMonitorInstance instance,
	struct CRHTradingAccountField *pRHTradingAccountField,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID, bool isLast
);

void cgoOnRspQryInvestorPosition
(
	CRHMonitorInstance instance,
	struct CRHMonitorPositionField *pRHMonitorPositionField,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID, bool isLast
);

void cgoOnRspOffsetOrder
(
	CRHMonitorInstance instance,
	struct CRHMonitorOffsetOrderField *pMonitorOrderField,
	struct CRHRspInfoField *pRHRspInfoField,
	int nRequestID, bool isLast
);

void cgoOnRtnOrder
(
	CRHMonitorInstance instance, struct CRHOrderField *pOrder
);

void cgoOnRtnTrade
(
	CRHMonitorInstance instance, struct CRHTradeField *pTrade
);

void cgoOnRtnInvestorMoney
(
	CRHMonitorInstance instance,
	struct CRHTradingAccountField *pRohonTradingAccountField
);

void cgoOnRtnInvestorPosition
(
	CRHMonitorInstance instance,
	struct CRHMonitorPositionField *pRohonMonitorPositionField
);

*/
import "C"

import (
	"github.com/pkg/errors"
)

var (
	spiCache = make(map[C.CRHMonitorInstance]RHRiskSpi)

	ErrInstanceNotExist = errors.New("spi instance not found")

	callbacks = C.callback_t{
		cOnFrontConnected:         C.CbOnFrontConnected(C.cgoOnFrontConnected),
		cOnFrontDisconnected:      C.CbOnFrontDisconnected(C.cgoOnFrontDisconnected),
		cOnRspUserLogin:           C.CbOnRspUserLogin(C.cgoOnRspUserLogin),
		cOnRspUserLogout:          C.CbOnRspUserLogout(C.cgoOnRspUserLogout),
		cOnRspQryMonitorAccounts:  C.CbOnRspQryMonitorAccounts(C.cgoOnRspQryMonitorAccounts),
		cOnRspQryInvestorMoney:    C.CbOnRspQryInvestorMoney(C.cgoOnRspQryInvestorMoney),
		cOnRspQryInvestorPosition: C.CbOnRspQryInvestorPosition(C.cgoOnRspQryInvestorPosition),
		cOnRspOffsetOrder:         C.CbOnRspOffsetOrder(C.cgoOnRspOffsetOrder),
		cOnRtnOrder:               C.CbOnRtnOrder(C.cgoOnRtnOrder),
		cOnRtnTrade:               C.CbOnRtnTrade(C.cgoOnRtnTrade),
		cOnRtnInvestorMoney:       C.CbOnRtnInvestorMoney(C.cgoOnRtnInvestorMoney),
		cOnRtnInvestorPosition:    C.CbOnRtnInvestorPosition(C.cgoOnRtnInvestorPosition),
	}
)

type RHRiskSpi interface {
	OnFrontConnected()
	OnFrontDisconnected(reason Reason)
	OnRspUserLogin(*RspUserLogin, *RspInfo, int64)
	OnRspUserLogout(*RspUserLogout, *RspInfo, int64)
	OnRspQryMonitorAccounts(*Investor, *RspInfo, int64, bool)
	OnRspQryInvestorMoney(*Account, *RspInfo, int64, bool)
	OnRspQryInvestorPosition(*Position, *RspInfo, int64, bool)
	OnRspOffsetOrder(*OffsetOrder, *RspInfo, int64, bool)
	OnRtnOrder(*Order)
	OnRtnTrade(*Trade)
	OnRtnInvestorMoney(*Account)
	OnRtnInvestorPosition(*Position)
}

func getSpiInstance(instance C.CRHMonitorInstance) (api RHRiskSpi) {
	var exist bool

	if api, exist = spiCache[instance]; !exist || api == nil {
		panic(errors.WithStack(ErrInstanceNotExist))
	}

	return
}

//export cgoOnFrontConnected
func cgoOnFrontConnected(instance C.CRHMonitorInstance) {
	getSpiInstance(instance).OnFrontConnected()
}

//export cgoOnFrontDisconnected
func cgoOnFrontDisconnected(instance C.CRHMonitorInstance, nReason C.int) {
	getSpiInstance(instance).OnFrontDisconnected(Reason(nReason))
}

//export cgoOnRspUserLogin
func cgoOnRspUserLogin(
	instance C.CRHMonitorInstance,
	pRspUserLoginField *C.struct_CRHMonitorRspUserLoginField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int,
) {
	login := NewFromCRHMonitorRspUserLoginField(pRspUserLoginField)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspUserLogin(login, info, int64(nRequestID))
}

//export cgoOnRspUserLogout
func cgoOnRspUserLogout(
	instance C.CRHMonitorInstance,
	pRspUserLoginField *C.struct_CRHMonitorUserLogoutField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int) {
	logout := NewFromCRHMonitorUserLogoutField(pRspUserLoginField)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspUserLogout(logout, info, int64(nRequestID))
}

//export cgoOnRspQryMonitorAccounts
func cgoOnRspQryMonitorAccounts(
	instance C.CRHMonitorInstance,
	pRspMonitorUser *C.struct_CRHQryInvestorField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int, isLast C.bool,
) {
	investor := NewFromCRHQryInvestorField(pRspMonitorUser)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspQryMonitorAccounts(
		investor, info, int64(nRequestID), bool(isLast),
	)
}

//export cgoOnRspQryInvestorMoney
func cgoOnRspQryInvestorMoney(
	instance C.CRHMonitorInstance,
	pRHTradingAccountField *C.struct_CRHTradingAccountField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int, isLast C.bool,
) {
	account := NewFromCRHTradingAccountField(pRHTradingAccountField)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspQryInvestorMoney(
		account, info, int64(nRequestID), bool(isLast),
	)
}

//export cgoOnRspQryInvestorPosition
func cgoOnRspQryInvestorPosition(
	instance C.CRHMonitorInstance,
	pRHMonitorPositionField *C.struct_CRHMonitorPositionField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int, isLast C.bool,
) {
	pos := NewFromCRHMonitorPositionField(pRHMonitorPositionField)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspQryInvestorPosition(
		pos, info, int64(nRequestID), bool(isLast),
	)
}

//export cgoOnRspOffsetOrder
func cgoOnRspOffsetOrder(
	instance C.CRHMonitorInstance,
	pMonitorOrderField *C.struct_CRHMonitorOffsetOrderField,
	pRHRspInfoField *C.struct_CRHRspInfoField,
	nRequestID C.int, isLast C.bool,
) {
	offsetOrd := NewFromCRHMonitorOffsetOrderField(pMonitorOrderField)
	info := NewFromCRHRspInfoField(pRHRspInfoField)

	getSpiInstance(instance).OnRspOffsetOrder(
		offsetOrd, info, int64(nRequestID), bool(isLast),
	)
}

//export cgoOnRtnOrder
func cgoOnRtnOrder(
	instance C.CRHMonitorInstance, pOrder *C.struct_CRHOrderField,
) {
	ord := NewFromCRHOrderField(pOrder)

	getSpiInstance(instance).OnRtnOrder(ord)
}

//export cgoOnRtnTrade
func cgoOnRtnTrade(
	instance C.CRHMonitorInstance, pTrade *C.struct_CRHTradeField,
) {
	td := NewFromCRHTradeField(pTrade)

	getSpiInstance(instance).OnRtnTrade(td)
}

//export cgoOnRtnInvestorMoney
func cgoOnRtnInvestorMoney(
	instance C.CRHMonitorInstance,
	pRohonTradingAccountField *C.struct_CRHTradingAccountField,
) {
	acct := NewFromCRHTradingAccountField(pRohonTradingAccountField)

	getSpiInstance(instance).OnRtnInvestorMoney(acct)
}

//export cgoOnRtnInvestorPosition
func cgoOnRtnInvestorPosition(
	instance C.CRHMonitorInstance,
	pRohonMonitorPositionField *C.struct_CRHMonitorPositionField,
) {
	pos := NewFromCRHMonitorPositionField(pRohonMonitorPositionField)

	getSpiInstance(instance).OnRtnInvestorPosition(pos)
}

func RegisterRHRiskSpi(instance C.CRHMonitorInstance, spi RHRiskSpi) {
	if _, exist := spiCache[instance]; !exist {
		C.SetCallbacks(instance, &callbacks)
		spiCache[instance] = spi
	}
}

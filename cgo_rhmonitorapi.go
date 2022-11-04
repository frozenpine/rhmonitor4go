package rhmonitor4go

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR}/libs -lcRHMonitorApi

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

func CheckRspInfo(info *RspInfo) error {
	if info == nil {
		return nil
	}

	if info.ErrorID != 0 {
		return errors.Errorf("ERROR[%d] %s", info.ErrorID, info.ErrorMsg)
	}

	return nil
}

func printData[T Investor | Account | Order | Trade | Position | OffsetOrder](inter string, data *T) {
	if data != nil {
		fmt.Printf("%s %s: %+v\n", inter, reflect.TypeOf(data), data)
	}
}

type RHMonitorApi struct {
	initOnce      sync.Once
	releaseOnce   sync.Once
	cInstance     C.CRHMonitorInstance
	brokerID      string
	remoteAddr    net.IP
	remotePort    int
	riskUser      RiskUser
	isConnected   atomic.Bool
	isLogin       atomic.Bool
	investorReady atomic.Bool
	investors     []*Investor
	requestID     int64
}

func (api *RHMonitorApi) nextRequestID() int {
	return int(atomic.AddInt64(&api.requestID, 1))
}

func (api *RHMonitorApi) waitBool(flag *atomic.Bool, v bool) {
	for !flag.CompareAndSwap(v, v) {
	}
}

func (api *RHMonitorApi) ReqUserLogin(login *RiskUser) int {
	if login != nil {
		api.riskUser = *login
	} else {
		log.Printf("No risk user info.")
		return -255
	}

	api.waitBool(&api.isConnected, true)

	return int(C.ReqUserLogin(
		api.cInstance,
		api.riskUser.ToCRHMonitorReqUserLoginField(),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqUserLogout() int {
	return int(C.ReqUserLogout(
		api.cInstance,
		api.riskUser.ToCRHMonitorUserLogoutField(),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqQryMonitorAccounts() int {
	api.waitBool(&api.isLogin, true)

	return int(C.ReqQryMonitorAccounts(
		api.cInstance,
		api.riskUser.ToCRHMonitorQryMonitorUser(),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqQryInvestorMoney(investor *Investor) int {
	api.waitBool(&api.isLogin, true)

	return int(C.ReqQryInvestorMoney(
		api.cInstance,
		investor.ToCRHMonitorQryInvestorMoneyField(),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqQryAllInvestorMoney() (rtn int) {
	api.waitBool(&api.investorReady, true)

	for _, investor := range api.investors {
		rtn = api.ReqQryInvestorMoney(investor)

		if rtn != 0 {
			log.Printf("Qeury money for user[%s] failed: %d", investor.InvestorID, rtn)
			break
		}
	}

	return
}

func (api *RHMonitorApi) ReqQryInvestorPosition(investor *Investor, instrumentID string) int {
	api.waitBool(&api.isLogin, true)

	return int(C.ReqQryInvestorPosition(
		api.cInstance,
		investor.ToCRHMonitorQryInvestorPositionField(instrumentID),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqQryAllInvestorPosition() (rtn int) {
	api.waitBool(&api.investorReady, true)

	for _, investor := range api.investors {
		rtn = api.ReqQryInvestorPosition(investor, "")

		if rtn != 0 {
			log.Printf("Qeury postion for user[%s] failed: %d", investor.InvestorID, rtn)
			break
		}
	}

	return
}

func (api *RHMonitorApi) ReqOffsetOrder(offsetOrder *OffsetOrder) int {
	log.Printf("ReqOffsetOrder not implied: %+v", offsetOrder)
	return -255
}

func (api *RHMonitorApi) ReqSubPushInfo(sub *SubInfo) int {
	api.waitBool(&api.isLogin, true)

	return int(C.ReqSubPushInfo(
		api.cInstance,
		sub.ToCRHMonitorSubPushInfo(),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqSubInvestorOrder(investor *Investor) int {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Order

	return api.ReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubAllInvestorOrder() (rtn int) {
	api.waitBool(&api.investorReady, true)

	for _, investor := range api.investors {
		rtn = api.ReqSubInvestorOrder(investor)

		if rtn != 0 {
			log.Printf("Sub investor[%s]'s order failed: %d", investor.InvestorID, rtn)
			break
		}
	}

	return
}

func (api *RHMonitorApi) ReqSubInvestorTrade(investor *Investor) int {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Trade

	return api.ReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubAllInvestorTrade() (rtn int) {
	api.waitBool(&api.investorReady, true)

	for _, investor := range api.investors {
		rtn = api.ReqSubInvestorTrade(investor)

		if rtn != 0 {
			log.Printf("Sub investor[%s]'s trade failed: %d", investor.InvestorID, rtn)
			break
		}
	}

	return
}

func (api *RHMonitorApi) OnFrontConnected() {
	log.Printf("Rohon risk[%s:%d] connected.", api.remoteAddr, api.remotePort)
	api.isConnected.CompareAndSwap(false, true)

	if api.isLogin.Load() && api.riskUser.IsValid() {
		api.ReqUserLogin(&api.riskUser)
	}
}

func (api *RHMonitorApi) OnFrontDisconnected(reason Reason) {
	log.Printf("Rohon risk[%s:%d] disconnected: %s", api.remoteAddr, api.remotePort, reason)
	api.isConnected.CompareAndSwap(true, false)
}

func (api *RHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, requestID int) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User[%s] login failed: %v", api.riskUser.UserID, err)
		return
	}

	if login != nil {
		log.Printf("User[%s] logged in: %s %s", api.riskUser.UserID, login.TradingDay, login.LoginTime)

		api.isLogin.CompareAndSwap(false, true)
	} else {
		log.Print("User login response data is nil.")
	}
}

func (api *RHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, requestID int) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User logout failed: %v", err)
		return
	}

	if logout != nil {
		log.Printf("User[%s] logged out.", logout.UserID)
		api.isLogin.CompareAndSwap(true, false)
	} else {
		log.Print("User logout response data is nil.")
	}
}

func (api *RHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, requestID int, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Monitor account query failed: %v", err)
		return
	}

	api.investors = append(api.investors, investor)

	if isLast {
		log.Printf("All monitor account query finished: %d", len(api.investors))

		for _, investor := range api.investors {
			printData("OnRspQryMonitorAccounts", investor)
		}

		api.investorReady.CompareAndSwap(false, true)
	}
}

func (api *RHMonitorApi) OnRspQryInvestorMoney(account *Account, info *RspInfo, requestID int, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Investor money query failed: %v", err)
		return
	}

	printData("OnRspQryInvestorMoney", account)
}

func (api *RHMonitorApi) OnRspQryInvestorPosition(position *Position, info *RspInfo, requestID int, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Investor position query failed: %v", err)

		return
	}

	printData("OnRspQryInvestorPosition", position)

	if isLast {
		log.Printf("Query investor[%s]'s position finished.", position.InvestorID)
	}
}

func (api *RHMonitorApi) OnRspOffsetOrder(offsetOrd *OffsetOrder, info *RspInfo, requestID int, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf(
			"Offset order[%s %s: %d@%f] for investor[%s] failed: %v",
			offsetOrd.Direction, offsetOrd.InstrumentID,
			offsetOrd.Volume, offsetOrd.Price,
			offsetOrd.InvestorID, err,
		)

		return
	}

	printData("OnRspOffsetOrder", offsetOrd)
}

func (api *RHMonitorApi) OnRtnOrder(order *Order) {
	printData("OnRtnOrder", order)
}

func (api *RHMonitorApi) OnRtnTrade(trade *Trade) {
	printData("OnRtnTrade", trade)
}

func (api *RHMonitorApi) OnRtnInvestorMoney(account *Account) {
	printData("OnRtnInvestorMoney", account)
}

func (api *RHMonitorApi) OnRtnInvestorPosition(position *Position) {
	printData("OnRtnInvestorPosition", position)
}

func NewRHMonitorApi(brokerID, addr string, port int) *RHMonitorApi {
	if addr == "" || port <= 1024 {
		log.Println("Rohon remote config is empty.")
		return nil
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		log.Printf("Invalid remote addr: %s", addr)
		return nil
	}

	cApi := C.CreateRHMonitorApi()

	api := RHMonitorApi{
		cInstance:  cApi,
		brokerID:   brokerID,
		remoteAddr: ip,
		remotePort: port,
	}

	C.SetCallbacks(cApi, &callbacks)

	instanceCache[cApi] = &api

	C.Init(cApi, C.CString(ip.String()), C.uint(port))

	return &api
}

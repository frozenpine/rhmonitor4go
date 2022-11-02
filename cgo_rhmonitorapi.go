package rohon

/*
#cgo CFLAGS: -I${SRCDIR}/../cRHMonitorApi -I${SRCDIR}/../includes/rohon
#cgo LDFLAGS: -L${SRCDIR}/../libs/rohon -lcRHMonitorApi

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"log"
	"net"
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

func printData(info *RspInfo, data interface{}) {
	if info != nil {
		log.Printf("RSP: %v %v", info, data)
	} else {
		log.Printf("RTN: %v", data)
	}
}

type RHMonitorApi struct {
	initOnce    sync.Once
	releaseOnce sync.Once
	cInstance   C.CRHMonitorInstance
	brokerID    string
	remoteAddr  net.IP
	remotePort  int
	riskUser    RiskUser
	isConnected atomic.Bool
	isLogin     atomic.Bool
	investors   []*Investor
	requestID   int64
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

func (api *RHMonitorApi) ReqQryInvestorPosition(investor *Investor, instrumentID string) int {
	api.waitBool(&api.isLogin, true)

	return int(C.ReqQryInvestorPosition(
		api.cInstance,
		investor.ToCRHMonitorQryInvestorPositionField(instrumentID),
		C.int(api.nextRequestID()),
	))
}

func (api *RHMonitorApi) ReqOffsetOrder(offsetOrder *OffsetOrder) int {
	log.Printf("ReqOffsetOrder not implied: %v", offsetOrder)
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

func (api *RHMonitorApi) OnFrontConnected() {
	log.Printf("Rohon risk[%s:%d] connected.", api.remoteAddr, api.remotePort)
	api.isConnected.CompareAndSwap(false, true)
}

func (api *RHMonitorApi) OnFrontDisconnected(reason Reason) {
	log.Printf("Rohon risk[%s:%d] disconnected: %v", api.remoteAddr, api.remotePort, reason)
	api.isConnected.CompareAndSwap(true, false)
}

func (api *RHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, requestID int) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User[%s] login failed: %v", api.riskUser.UserID, err)
		return
	}

	log.Printf("User[%s] logged in: %s %s", login.UserID, login.TradingDay, login.LoginTime)
}

func (api *RHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, requestID int) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User logout failed: %v", err)
		return
	}

	log.Printf("User[%s] logged out.", logout.UserID)
}

func (api *RHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, requestID int, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Monitor account query failed: %v", err)
		return
	}

	api.investors = append(api.investors, investor)

	if isLast {
		log.Printf("All monitor account query finished: %v", api.investors)
	}
}

func (api *RHMonitorApi) OnRspQryInvestorMoney(account *Account, info *RspInfo, requestID int, isLast bool) {
	printData(info, account)
}

func (api *RHMonitorApi) OnRspQryInvestorPosition(position *Position, info *RspInfo, requestID int, isLast bool) {
	printData(info, position)
}

func (api *RHMonitorApi) OnRspOffsetOrder(offsetOrd *OffsetOrder, info *RspInfo, requestID int, isLast bool) {
	printData(info, offsetOrd)
}

func (api *RHMonitorApi) OnRtnOrder(order *Order) {
	printData(nil, order)
}

func (api *RHMonitorApi) OnRtnTrade(trade *Trade) {
	printData(nil, trade)
}

func (api *RHMonitorApi) OnRtnInvestorMoney(account *Account) {
	printData(nil, account)
}

func (api *RHMonitorApi) OnRtnInvestorPosition(position *Position) {
	printData(nil, position)
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

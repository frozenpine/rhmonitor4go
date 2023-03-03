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

var (
	OrderTopic     = "order"
	OffsetOrdTopic = "offsetOrd"
	TradeTopic     = "trade"
	PositionTopic  = "position"
	MoneyTopic     = "money"
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

type RHRiskApi interface {
	ReqUserLogin(*RiskUser) (int64, int)
	ReqUserLogout() (int64, int)
	ReqQryMonitorAccounts() (int64, int)
	ReqQryInvestorMoney(*Investor) (int64, int)
	ReqQryInvestorPosition(*Investor, string) (int64, int)
	ReqOffsetOrder(*OffsetOrder) (int64, int)
	ReqSubPushInfo(*SubInfo) (int64, int)
	ReqSubInvestorOrder(*Investor) (int64, int)
	ReqSubInvestorTrade(*Investor) (int64, int)
}

type RHMonitorApi struct {
	initOnce    sync.Once
	releaseOnce sync.Once
	cInstance   C.CRHMonitorInstance

	brokerID   string
	remoteAddr net.IP
	remotePort int
	riskUser   RiskUser

	requests  *RequestCache
	investors *InvestorCache
	accounts  *AccountCache

	requestID int64
}

func (api *RHMonitorApi) nextRequestID(out *int64) int {
	var old int64

	for {
		old = api.requestID
		if atomic.CompareAndSwapInt64(&api.requestID, old, old+1) {
			break
		}
	}

	*out = old + 1
	return int(old + 1)
}

func (api *RHMonitorApi) ReqUserLogin(login *RiskUser) (int64, int) {
	var reqID int64

	if login != nil {
		api.riskUser = *login
	} else {
		log.Printf("No risk user info.")
		return -1, -255
	}

	return reqID, api.requests.WaitConnectedAndDo(func() int {
		log.Printf("Request user login with cache: %+v", login)

		return int(C.ReqUserLogin(
			api.cInstance,
			api.riskUser.ToCRHMonitorReqUserLoginField(),
			C.int(api.nextRequestID(&reqID)),
		))
	})
}

func (api *RHMonitorApi) ReqUserLogout() (int64, int) {
	var reqID int64

	log.Printf("Request user logout: %+v", api.riskUser)

	return reqID, int(C.ReqUserLogout(
		api.cInstance,
		api.riskUser.ToCRHMonitorUserLogoutField(),
		C.int(api.nextRequestID(&reqID)),
	))
}

func (api *RHMonitorApi) ReqQryMonitorAccounts() (int64, int) {
	var reqID int64

	return reqID, api.requests.WaitLoginAndDo(func() int {
		log.Println("Request query monitor accounts with cache.")

		return int(C.ReqQryMonitorAccounts(
			api.cInstance,
			api.riskUser.ToCRHMonitorQryMonitorUser(),
			C.int(api.nextRequestID(&reqID)),
		))
	})
}

func (api *RHMonitorApi) ExecReqQryInvestorMoney(investor *Investor) (int64, int) {
	var reqID int64

	api.requests.WaitLogin()

	log.Printf("Request query investor's money w/o cache: %+v", investor)

	return reqID, int(C.ReqQryInvestorMoney(
		api.cInstance,
		investor.ToCRHMonitorQryInvestorMoneyField(),
		C.int(api.nextRequestID(&reqID)),
	))
}

func (api *RHMonitorApi) ReqQryInvestorMoney(investor *Investor) (int64, int) {
	var reqID int64

	return reqID, api.requests.WaitLoginAndDo(func() int {
		log.Printf("Request query investor's money with cache: %+v", investor)

		return int(C.ReqQryInvestorMoney(
			api.cInstance,
			investor.ToCRHMonitorQryInvestorMoneyField(),
			C.int(api.nextRequestID(&reqID)),
		))
	})
}

func (api *RHMonitorApi) ReqQryAllInvestorMoney() int {
	return api.requests.WaitInvestorReadyAndDo(func() (rtn int) {
		log.Printf("Request query all investor's money with cache.")

		api.investors.ForEach(func(_ string, investor *Investor) bool {
			_, rtn = api.ExecReqQryInvestorMoney(investor)

			if rtn != 0 {
				log.Printf("Query money for user[%s] failed: %d", investor.InvestorID, rtn)
				return false
			}

			return true
		})
		return
	})
}

func (api *RHMonitorApi) ExecReqQryInvestorPosition(investor *Investor, instrumentID string) (int64, int) {
	var reqID int64

	api.requests.WaitLogin()

	log.Printf("Query investor's position w/o cache: %+v @ %s", investor, instrumentID)

	return reqID, int(C.ReqQryInvestorPosition(
		api.cInstance,
		investor.ToCRHMonitorQryInvestorPositionField(instrumentID),
		C.int(api.nextRequestID(&reqID)),
	))
}

func (api *RHMonitorApi) ReqQryInvestorPosition(investor *Investor, instrumentID string) (int64, int) {
	var reqID int64

	return reqID, api.requests.WaitLoginAndDo(func() int {
		log.Printf("Query investor's position with cache: %+v @ %s", investor, instrumentID)

		return int(C.ReqQryInvestorPosition(
			api.cInstance,
			investor.ToCRHMonitorQryInvestorPositionField(instrumentID),
			C.int(api.nextRequestID(&reqID)),
		))
	})
}

func (api *RHMonitorApi) ReqQryAllInvestorPosition() int {
	return api.requests.WaitInvestorReadyAndDo(func() (rtn int) {
		log.Printf("Request query all investor's position with cache.")

		api.investors.ForEach(func(_ string, investor *Investor) bool {
			_, rtn = api.ExecReqQryInvestorPosition(investor, "")

			if rtn != 0 {
				log.Printf("Qeury postion for user[%s] failed: %d", investor.InvestorID, rtn)
				return false
			}

			return true
		})

		return
	})
}

func (api *RHMonitorApi) ReqOffsetOrder(offsetOrder *OffsetOrder) (int64, int) {
	log.Printf("ReqOffsetOrder not implied: %+v", offsetOrder)
	return -1, -255
}

func (api *RHMonitorApi) ExecReqSubPushInfo(sub *SubInfo) (int64, int) {
	var reqID int64

	api.requests.WaitLogin()

	log.Printf("Request sub info w/o cache: %+v", sub)

	return reqID, int(C.ReqSubPushInfo(
		api.cInstance,
		sub.ToCRHMonitorSubPushInfo(),
		C.int(api.nextRequestID(&reqID)),
	))
}

func (api *RHMonitorApi) ReqSubPushInfo(sub *SubInfo) (int64, int) {
	var reqID int64

	return reqID, api.requests.WaitLoginAndDo(func() int {
		log.Printf("Request sub info with cache: %+v", sub)

		return int(C.ReqSubPushInfo(
			api.cInstance,
			sub.ToCRHMonitorSubPushInfo(),
			C.int(api.nextRequestID(&reqID)),
		))
	})
}

func (api *RHMonitorApi) ExecReqSubInvestorOrder(investor *Investor) (int64, int) {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Order

	return api.ExecReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubInvestorOrder(investor *Investor) (int64, int) {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Order

	return api.ReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubAllInvestorOrder() int {
	return api.requests.WaitInvestorReadyAndDo(func() (rtn int) {
		log.Printf("Request sub all investor[%d]'s order with cache.", api.investors.Size())

		api.investors.ForEach(func(_ string, investor *Investor) bool {
			_, rtn = api.ExecReqSubInvestorOrder(investor)

			if rtn != 0 {
				log.Printf("Sub investor[%s]'s order failed: %d", investor.InvestorID, rtn)
				return false
			}

			return true
		})

		return
	})

}

func (api *RHMonitorApi) ExecReqSubInvestorTrade(investor *Investor) (int64, int) {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Trade

	return api.ExecReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubInvestorTrade(investor *Investor) (int64, int) {
	sub := SubInfo{}
	sub.BrokerID = investor.BrokerID
	sub.InvestorID = investor.InvestorID
	sub.AccountType = RH_ACCOUNTTYPE_VIRTUAL
	sub.SubInfoType = RHMonitorSubPushInfoType_Trade

	return api.ReqSubPushInfo(&sub)
}

func (api *RHMonitorApi) ReqSubAllInvestorTrade() int {
	return api.requests.WaitInvestorReadyAndDo(func() (rtn int) {
		log.Printf("Request sub all investor[%d]'s trade info with cache.", api.investors.Size())

		api.investors.ForEach(func(_ string, investor *Investor) bool {
			_, rtn = api.ExecReqSubInvestorTrade(investor)

			if rtn != 0 {
				log.Printf("Sub investor[%s]'s trade failed: %d", investor.InvestorID, rtn)
				return false
			}

			return true
		})

		return
	})
}

func (api *RHMonitorApi) HandleConnected() {
	api.requests.SetConnected(true)

	if rtn := api.requests.RedoConnected(); rtn != 0 {
		log.Printf("Redo connected failed with error: %d", rtn)
	}
}

func (api *RHMonitorApi) OnFrontConnected() {
	api.HandleConnected()
	log.Printf("Rohon risk[%s:%d] connected.", api.remoteAddr, api.remotePort)
}

func (api *RHMonitorApi) HandleDisconnected() {
	api.requests.SetInvestorReady(false)
	api.requests.SetLogin(false)
	api.requests.SetConnected(false)
}

func (api *RHMonitorApi) OnFrontDisconnected(reason Reason) {
	api.HandleDisconnected()
	log.Printf("Rohon risk[%s:%d] disconnected: %s", api.remoteAddr, api.remotePort, reason)
}

func (api *RHMonitorApi) HandleLogin(login *RspUserLogin) {
	if login != nil {
		api.requests.SetLogin(true)

		if rtn := api.requests.RedoLoggedIn(); rtn != 0 {
			log.Printf("Redo login failed with error: %d", rtn)
		}
	} else {
		log.Print("User login response data is nil.")
	}
}

func (api *RHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, requestID int64) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User[%s] login failed: %v", api.riskUser.UserID, err)
		return
	}

	api.HandleLogin(login)

	if login != nil {
		log.Printf("User[%s] logged in: %s %s", api.riskUser.UserID, login.TradingDay, login.LoginTime)
	}
}

func (api *RHMonitorApi) HandleLogout(logout *RspUserLogout) {
	if logout != nil {
		api.requests.SetLogin(false)
	} else {
		log.Print("User logout response data is nil.")
	}
}

func (api *RHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, requestID int64) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("User logout failed: %v", err)
		return
	}

	if logout != nil {
		log.Printf("User[%s] logged out.", logout.UserID)
	}
}

func (api *RHMonitorApi) HandleInvestor(investor *Investor, isLast bool) {
	if investor != nil {
		api.investors.AddInvestor(investor)
	}

	if isLast {
		api.requests.SetInvestorReady(true)

		api.requests.RedoInvestorReady()
	}
}

func (api *RHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, requestID int64, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Monitor account query failed: %v", err)
		return
	}

	api.HandleInvestor(investor, isLast)

	if isLast {
		log.Printf("All monitor account query finished: %d", api.investors.Size())

		api.investors.ForEach(func(_ string, investor *Investor) bool {
			printData("OnRspQryMonitorAccounts", investor)
			return true
		})
	}
}

func (api *RHMonitorApi) OnRspQryInvestorMoney(account *Account, info *RspInfo, requestID int64, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Investor money query failed: %v", err)
		return
	}

	printData("OnRspQryInvestorMoney", account)
}

func (api *RHMonitorApi) OnRspQryInvestorPosition(position *Position, info *RspInfo, requestID int64, isLast bool) {
	if err := CheckRspInfo(info); err != nil {
		log.Printf("Investor position query failed: %v", err)

		return
	}

	printData("OnRspQryInvestorPosition", position)

	if isLast {
		log.Printf("Query investor[%s]'s position finished.", position.InvestorID)
	}
}

func (api *RHMonitorApi) OnRspOffsetOrder(offsetOrd *OffsetOrder, info *RspInfo, requestID int64, isLast bool) {
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

// 不进行任何订阅，在进行用户资金查询或其他任何会影响资金变动的操作后，都会推送
func (api *RHMonitorApi) OnRtnInvestorMoney(account *Account) {
	printData("OnRtnInvestorMoney", account)
}

func (api *RHMonitorApi) OnRtnInvestorPosition(position *Position) {
	printData("OnRtnInvestorPosition", position)
}

func (api *RHMonitorApi) Init(brokerID, addr string, port int, spi RHRiskSpi) error {
	if addr == "" || port <= 1024 {
		return errors.New("rohon remote config is empty")
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return errors.Errorf("invalid remote addr: %s", addr)
	}

	api.initOnce.Do(func() {
		api.cInstance = C.CreateRHMonitorApi()
		api.brokerID = brokerID
		api.remoteAddr = ip
		api.remotePort = port
		api.requests = &RequestCache{}
		api.investors = &InvestorCache{
			data: make(map[string]*Investor),
		}

		if spi == nil {
			spi = api
		}
		RegisterRHRiskSpi(api.cInstance, spi)

		C.Init(api.cInstance, C.CString(ip.String()), C.uint(port))
	})

	return nil
}

func NewRHMonitorApi(brokerID, addr string, port int) *RHMonitorApi {
	api := RHMonitorApi{}

	if err := api.Init(brokerID, addr, port, nil); err != nil {
		log.Printf("Create RHMonitorApi failed: %+v", err)
		return nil
	}

	return &api
}

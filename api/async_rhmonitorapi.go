package api

import (
	"log"
	"sync"

	"github.com/frozenpine/rhmonitor4go"
)

type RHRiskData interface {
	rhmonitor4go.RspUserLogin | rhmonitor4go.RspUserLogout |
		rhmonitor4go.Investor | rhmonitor4go.Account |
		rhmonitor4go.Position | rhmonitor4go.OffsetOrder |
		rhmonitor4go.Order | rhmonitor4go.Trade
}

type CallbackFn[T RHRiskData] func(Result[T]) error

func promiseMissingHandler(reqID int64, info *rhmonitor4go.RspInfo, data interface{}) {
	log.Printf("No promise found for req[%d]: %s\n\t%+v", reqID, info, data)
}

type AsyncRHMonitorApi struct {
	RHMonitorApi

	promiseCache sync.Map
	flowCache    sync.Map
}

func NewAsyncRHMonitorApi(brokerID, addr string, port int) *AsyncRHMonitorApi {
	api := AsyncRHMonitorApi{}

	if err := api.Init(brokerID, addr, port, &api); err != nil {
		log.Printf("Create AsyncRHMonitorApi failed: %+v", err)
		return nil
	}

	return &api
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogin(login *rhmonitor4go.RiskUser) Result[rhmonitor4go.RspUserLogin] {
	reqID, rtn := api.ReqUserLogin(login)

	result := NewSingleResult[rhmonitor4go.RspUserLogin](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogin(
	login *rhmonitor4go.RspUserLogin, info *rhmonitor4go.RspInfo, reqID int64,
) {
	api.HandleLogin(login)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(Result[rhmonitor4go.RspUserLogin])

		result.AppendResult(reqID, login, info, true)
	} else {
		promiseMissingHandler(reqID, info, login)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogout() Result[rhmonitor4go.RspUserLogout] {
	reqID, rtn := api.ReqUserLogout()
	result := NewSingleResult[rhmonitor4go.RspUserLogout](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogout(
	logout *rhmonitor4go.RspUserLogout, info *rhmonitor4go.RspInfo, reqID int64,
) {
	api.HandleLogout(logout)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(Result[rhmonitor4go.RspUserLogout])

		result.AppendResult(reqID, logout, info, true)
	} else {
		promiseMissingHandler(reqID, info, logout)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryMonitorAccounts() Result[rhmonitor4go.Investor] {
	reqID, rtn := api.ReqQryMonitorAccounts()

	result := NewSingleResult[rhmonitor4go.Investor](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryMonitorAccounts(
	investor *rhmonitor4go.Investor,
	info *rhmonitor4go.RspInfo,
	reqID int64, isLast bool,
) {
	api.HandleInvestor(investor, isLast)

	if promise, exist := api.promiseCache.Load(reqID); exist {
		result := promise.(Result[rhmonitor4go.Investor])

		result.AppendResult(reqID, investor, info, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	} else {
		promiseMissingHandler(reqID, info, investor)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorMoney(investor *rhmonitor4go.Investor) Result[rhmonitor4go.Account] {
	reqID, rtn := api.ReqQryInvestorMoney(investor)
	result := NewSingleResult[rhmonitor4go.Account](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) AsyncReqQryAllInvestorMoney() Result[rhmonitor4go.Account] {
	api.requests.WaitInvestorReady()

	results := NewBatchResult[rhmonitor4go.Account]()

	api.investors.ForEach(func(s string, i *rhmonitor4go.Investor) bool {
		reqID, rtn := api.ExecReqQryInvestorMoney(i)

		api.promiseCache.Store(reqID, results)
		results.AppendRequest(reqID, rtn)

		if rtn != 0 {
			log.Printf("Query money for user[%s] failed: %d", i.InvestorID, rtn)
			return false
		}

		return true
	})

	return results
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorMoney(acct *rhmonitor4go.Account, info *rhmonitor4go.RspInfo, reqID int64, isLast bool) {
	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[rhmonitor4go.Account])

		result.AppendResult(reqID, acct, info, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	} else {
		promiseMissingHandler(reqID, info, acct)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorPosition(investor *rhmonitor4go.Investor, instrument string) Result[rhmonitor4go.Position] {
	reqID, rtn := api.ReqQryInvestorPosition(investor, instrument)
	result := NewSingleResult[rhmonitor4go.Position](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorPosition(pos *rhmonitor4go.Position, info *rhmonitor4go.RspInfo, reqID int64, isLast bool) {
	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[rhmonitor4go.Position])

		result.AppendResult(reqID, pos, info, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	} else {
		promiseMissingHandler(reqID, info, pos)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqOffsetOrder(offset *rhmonitor4go.OffsetOrder) Result[rhmonitor4go.OffsetOrder] {
	reqID, rtn := api.ReqOffsetOrder(offset)
	result := NewSingleResult[rhmonitor4go.OffsetOrder](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspOffsetOrder(offset *rhmonitor4go.OffsetOrder, info *rhmonitor4go.RspInfo, reqID int64, isLast bool) {
	api.RHMonitorApi.OnRspOffsetOrder(offset, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[rhmonitor4go.OffsetOrder])

		result.AppendResult(reqID, offset, info, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	} else {
		promiseMissingHandler(reqID, info, offset)
	}
}

func (api *AsyncRHMonitorApi) GetOrderFlow() Result[rhmonitor4go.Order] {
	if flow, exist := api.flowCache.Load(orderFlow); exist {
		return flow.(Result[rhmonitor4go.Order])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetTradeFlow() Result[rhmonitor4go.Trade] {
	if flow, exist := api.flowCache.Load(tradeFlow); exist {
		return flow.(Result[rhmonitor4go.Trade])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetAccountFlow() Result[rhmonitor4go.Account] {
	if flow, exist := api.flowCache.Load(accountFlow); exist {
		return flow.(Result[rhmonitor4go.Account])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetPositionFlow() Result[rhmonitor4go.Position] {
	if flow, exist := api.flowCache.Load(positionFlow); exist {
		return flow.(Result[rhmonitor4go.Position])
	}

	return nil
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorOrder(investor *rhmonitor4go.Investor) Result[rhmonitor4go.Order] {
	api.ReqSubInvestorOrder(investor)

	api.flowCache.LoadOrStore(
		orderFlow, NewFlowResult[rhmonitor4go.Order](0),
	)

	return api.GetOrderFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorOrder() Result[rhmonitor4go.Order] {
	api.requests.WaitInvestorReady()

	api.flowCache.LoadOrStore(
		orderFlow, NewFlowResult[rhmonitor4go.Order](0),
	)

	api.investors.ForEach(func(s string, i *rhmonitor4go.Investor) bool {
		api.ExecReqSubInvestorOrder(i)

		return true
	})

	return api.GetOrderFlow()
}

func (api *AsyncRHMonitorApi) OnRtnOrder(order *rhmonitor4go.Order) {
	if flow := api.GetOrderFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, order, nil, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorTrade(investor *rhmonitor4go.Investor) Result[rhmonitor4go.Trade] {
	api.ReqSubInvestorTrade(investor)

	api.flowCache.LoadOrStore(
		tradeFlow, NewFlowResult[rhmonitor4go.Trade](0),
	)

	return api.GetTradeFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorTrade() Result[rhmonitor4go.Trade] {
	api.requests.WaitInvestorReady()

	api.flowCache.LoadOrStore(
		tradeFlow, NewFlowResult[rhmonitor4go.Trade](0),
	)

	api.investors.ForEach(func(s string, i *rhmonitor4go.Investor) bool {
		api.ExecReqSubInvestorTrade(i)

		return true
	})

	return api.GetTradeFlow()
}

func (api *AsyncRHMonitorApi) OnRtnTrade(trade *rhmonitor4go.Trade) {
	// api.RHMonitorApi.OnRtnTrade(trade)

	if flow := api.GetTradeFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, trade, nil, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorMoney() Result[rhmonitor4go.Account] {
	api.flowCache.LoadOrStore(
		accountFlow, NewFlowResult[rhmonitor4go.Account](0),
	)

	return api.GetAccountFlow()
}

func (api *AsyncRHMonitorApi) OnRtnInvestorMoney(account *rhmonitor4go.Account) {
	// api.RHMonitorApi.OnRtnInvestorMoney(account)

	if flow := api.GetAccountFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, account, nil, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorPosition(investor *rhmonitor4go.Investor) Result[rhmonitor4go.Position] {
	api.AsyncReqSubInvestorTrade(investor)

	api.flowCache.LoadOrStore(
		positionFlow, NewFlowResult[rhmonitor4go.Position](0),
	)

	return api.GetPositionFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorPosition() Result[rhmonitor4go.Position] {
	api.AsyncReqSubAllInvestorTrade()

	api.flowCache.LoadOrStore(
		positionFlow, NewFlowResult[rhmonitor4go.Position](0),
	)

	return api.GetPositionFlow()
}

func (api *AsyncRHMonitorApi) OnRtnInvestorPosition(position *rhmonitor4go.Position) {
	// api.RHMonitorApi.OnRtnInvestorPosition(position)

	if flow := api.GetPositionFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, position, nil, false)
	}
}

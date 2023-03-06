package rhmonitor4go

import (
	"log"
	"sync"
)

type RHRiskData interface {
	RspUserLogin | RspUserLogout | Investor | Account | Position | OffsetOrder | Order | Trade
}

type CallbackFn[T RHRiskData] func(Result[T]) error

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

func (api *AsyncRHMonitorApi) AsyncReqUserLogin(login *RiskUser) Result[RspUserLogin] {
	reqID, rtn := api.ReqUserLogin(login)

	result := NewSingleResult[RspUserLogin](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, reqID int64) {
	api.HandleLogin(login)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(Result[RspUserLogin])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, login, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogout() Result[RspUserLogout] {
	reqID, rtn := api.ReqUserLogout()
	result := NewSingleResult[RspUserLogout](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, reqID int64) {
	api.HandleLogout(logout)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(Result[RspUserLogout])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, logout, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryMonitorAccounts() Result[Investor] {
	reqID, rtn := api.ReqQryMonitorAccounts()

	result := NewSingleResult[Investor](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, reqID int64, isLast bool) {
	api.HandleInvestor(investor, isLast)

	if promise, exist := api.promiseCache.Load(reqID); exist {
		result := promise.(Result[Investor])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, investor, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorMoney(investor *Investor) Result[Account] {
	reqID, rtn := api.ReqQryInvestorMoney(investor)
	result := NewSingleResult[Account](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) AsyncReqQryAllInvestorMoney() Result[Account] {
	results := NewBatchResult[Account]()

	api.requests.WaitInvestorReady()

	api.investors.ForEach(func(s string, i *Investor) bool {
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

func (api *AsyncRHMonitorApi) OnRspQryInvestorMoney(acct *Account, info *RspInfo, reqID int64, isLast bool) {
	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[Account])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, acct, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorPosition(investor *Investor, instrument string) Result[Position] {
	reqID, rtn := api.ReqQryInvestorPosition(investor, instrument)
	result := NewSingleResult[Position](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorPosition(pos *Position, info *RspInfo, reqID int64, isLast bool) {
	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[Position])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, pos, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqOffsetOrder(offset *OffsetOrder) Result[OffsetOrder] {
	reqID, rtn := api.ReqOffsetOrder(offset)
	result := NewSingleResult[OffsetOrder](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspOffsetOrder(offset *OffsetOrder, info *RspInfo, reqID int64, isLast bool) {
	api.RHMonitorApi.OnRspOffsetOrder(offset, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(Result[OffsetOrder])

		result.SetRspInfo(reqID, info)

		result.AppendResult(reqID, offset, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) GetOrderFlow() Result[Order] {
	if flow, exist := api.flowCache.Load(orderFlow); exist {
		return flow.(Result[Order])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetTradeFlow() Result[Trade] {
	if flow, exist := api.flowCache.Load(tradeFlow); exist {
		return flow.(Result[Trade])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetAccountFlow() Result[Account] {
	if flow, exist := api.flowCache.Load(accountFlow); exist {
		return flow.(Result[Account])
	}

	return nil
}

func (api *AsyncRHMonitorApi) GetPositionFlow() Result[Position] {
	if flow, exist := api.flowCache.Load(positionFlow); exist {
		return flow.(Result[Position])
	}

	return nil
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorOrder(investor *Investor) Result[Order] {
	api.ReqSubInvestorOrder(investor)

	api.flowCache.LoadOrStore(
		orderFlow, NewFlowResult[Order](0),
	)

	return api.GetOrderFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorOrder() Result[Order] {
	api.requests.WaitInvestorReady()

	api.flowCache.LoadOrStore(
		orderFlow, NewFlowResult[Order](0),
	)

	api.investors.ForEach(func(s string, i *Investor) bool {
		api.ExecReqSubInvestorOrder(i)

		return true
	})

	return api.GetOrderFlow()
}

func (api *AsyncRHMonitorApi) OnRtnOrder(order *Order) {
	if flow := api.GetOrderFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, order, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorTrade(investor *Investor) Result[Trade] {
	api.ReqSubInvestorTrade(investor)

	api.flowCache.LoadOrStore(
		tradeFlow, NewFlowResult[Trade](0),
	)

	return api.GetTradeFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorTrade() Result[Trade] {
	api.requests.WaitInvestorReady()

	api.flowCache.LoadOrStore(
		tradeFlow, NewFlowResult[Trade](0),
	)

	api.investors.ForEach(func(s string, i *Investor) bool {
		api.ExecReqSubInvestorTrade(i)

		return true
	})

	return api.GetTradeFlow()
}

func (api *AsyncRHMonitorApi) OnRtnTrade(trade *Trade) {
	// api.RHMonitorApi.OnRtnTrade(trade)

	if flow := api.GetTradeFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, trade, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorMoney() Result[Account] {
	api.flowCache.LoadOrStore(
		accountFlow, NewFlowResult[Account](0),
	)

	return api.GetAccountFlow()
}

func (api *AsyncRHMonitorApi) OnRtnInvestorMoney(account *Account) {
	// api.RHMonitorApi.OnRtnInvestorMoney(account)

	if flow := api.GetAccountFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, account, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorPosition(investor *Investor) Result[Position] {
	api.AsyncReqSubInvestorTrade(investor)

	api.flowCache.LoadOrStore(
		positionFlow, NewFlowResult[Position](0),
	)

	return api.GetPositionFlow()
}

func (api *AsyncRHMonitorApi) AsyncReqSubAllInvestorPosition() Result[Position] {
	api.AsyncReqSubAllInvestorTrade()

	api.flowCache.LoadOrStore(
		positionFlow, NewFlowResult[Position](0),
	)

	return api.GetPositionFlow()
}

func (api *AsyncRHMonitorApi) OnRtnInvestorPosition(position *Position) {
	// api.RHMonitorApi.OnRtnInvestorPosition(position)

	if flow := api.GetPositionFlow(); flow != nil {
		flow.AppendResult(InfinitResultReq, position, false)
	}
}

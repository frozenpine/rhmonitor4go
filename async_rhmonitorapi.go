package rhmonitor4go

import (
	"log"
	"sync"
)

const (
	minDataLen = 1
	maxDataLen = 10
)

type RHRiskData interface {
	RspUserLogin | RspUserLogout | Investor | Account | Position | OffsetOrder | Order | Trade
}

type CallbackFn[T RHRiskData] func(*Result[T]) error

type Promise[T RHRiskData] interface {
	Await()
	Then(CallbackFn[T]) Promise[T]
	Catch(CallbackFn[T]) Promise[T]
	Finally(CallbackFn[T]) Promise[T]
}

type Result[T RHRiskData] struct {
	RequestID int
	ExecCode  int
	RspInfo   *RspInfo
	Data      chan *T

	start     chan struct{}
	startOnce sync.Once

	finish sync.WaitGroup

	notifyData chan struct{}
	notifyOnce sync.Once

	successFn CallbackFn[T]
	failFn    CallbackFn[T]
	finalFn   CallbackFn[T]
}

func NewResult[T RHRiskData](execCode int, dataLen int) *Result[T] {
	if dataLen < minDataLen {
		dataLen = minDataLen
	}

	if dataLen > maxDataLen {
		dataLen = maxDataLen
	}

	result := Result[T]{
		ExecCode:   execCode,
		start:      make(chan struct{}),
		notifyData: make(chan struct{}),
		Data:       make(chan *T, dataLen),
	}

	if execCode != 0 {
		close(result.notifyData)
		close(result.Data)
	}

	return &result
}

func (r *Result[T]) waitAsyncData() {
	<-r.start
	<-r.notifyData
}

func (r *Result[T]) AppendResult(v *T, isLast bool) {
	r.notifyOnce.Do(func() { close(r.notifyData) })

	r.Data <- v

	if isLast {
		close(r.Data)
	}
}

func (r *Result[T]) Await() {
	r.finish.Add(1)

	r.startOnce.Do(func() { close(r.start) })

	go func() {
		r.waitAsyncData()

		defer r.finish.Done()

		if r.ExecCode != 0 {
			return
		}

		var errors struct {
			thenErr  error
			catchErr error
			finalErr error
		}

		if r.RspInfo == nil || r.RspInfo.ErrorID == 0 {
			if r.successFn != nil {
				errors.thenErr = r.successFn(r)
			}
		} else if r.RspInfo.ErrorID != 0 {
			errors.catchErr = r.failFn(r)
		}

		if r.finalFn != nil {
			errors.finalErr = r.finalFn(r)
		}

		if errors.thenErr != nil {
			log.Printf("Async[Then] callback for request[%d] error: %+v", r.RequestID, errors.thenErr)
		}

		if errors.catchErr != nil {
			log.Printf("Async[Catch] callback for request[%d] error: %+v", r.RequestID, errors.thenErr)
		}

		if errors.finalErr != nil {
			log.Printf("Async[Finally] callback for request[%d] error: %+v", r.RequestID, errors.finalErr)
		}
	}()

	r.finish.Wait()
}

func (r *Result[T]) Then(fn CallbackFn[T]) Promise[T] {
	r.successFn = fn

	return r
}

func (r *Result[T]) Catch(fn CallbackFn[T]) Promise[T] {
	r.failFn = fn

	return r
}

func (r *Result[T]) Finally(fn CallbackFn[T]) Promise[T] {
	r.finalFn = fn

	return r
}

type AsyncRHMonitorApi struct {
	RHMonitorApi

	asyncCache sync.Map
}

func NewAsyncRHMonitorApi(brokerID, addr string, port int) *AsyncRHMonitorApi {
	api := AsyncRHMonitorApi{}

	if err := api.Init(brokerID, addr, port, &api); err != nil {
		log.Printf("Create AsyncRHMonitorApi failed: %+v", err)
		return nil
	}

	return &api
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogin(login *RiskUser) Promise[RspUserLogin] {
	reqID, rtn := api.ReqUserLogin(login)

	result := NewResult[RspUserLogin](rtn, 1)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogin(login, info, reqID)

	if async, exist := api.asyncCache.LoadAndDelete(reqID); exist {
		result := async.(*Result[RspUserLogin])

		result.RspInfo = info

		result.AppendResult(login, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogout() Promise[RspUserLogout] {
	reqID, rtn := api.ReqUserLogout()
	result := NewResult[RspUserLogout](rtn, 1)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogout(logout, info, reqID)

	if async, exist := api.asyncCache.LoadAndDelete(reqID); exist {
		result := async.(*Result[RspUserLogout])

		result.RspInfo = info

		result.AppendResult(logout, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryMonitorAccounts() Promise[Investor] {
	reqID, rtn := api.ReqQryMonitorAccounts()
	result := NewResult[Investor](rtn, maxDataLen)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryMonitorAccounts(investor, info, reqID, isLast)

	if async, exist := api.asyncCache.Load(reqID); exist {
		result := async.(*Result[Investor])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(investor, isLast)

		if isLast {
			api.asyncCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorMoney(investor *Investor) Promise[Account] {
	reqID, rtn := api.ReqQryInvestorMoney(investor)
	result := NewResult[Account](rtn, maxDataLen)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorMoney(acct *Account, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorMoney(acct, info, reqID, isLast)

	if async, exist := api.asyncCache.Load(reqID); exist {
		result := async.(*Result[Account])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(acct, isLast)

		if isLast {
			api.asyncCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorPosition(investor *Investor, instrument string) Promise[Position] {
	reqID, rtn := api.ReqQryInvestorPosition(investor, instrument)
	result := NewResult[Position](rtn, maxDataLen)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorPosition(pos *Position, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorPosition(pos, info, reqID, isLast)

	if async, exist := api.asyncCache.Load(reqID); exist {
		result := async.(*Result[Position])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(pos, isLast)

		if isLast {
			api.asyncCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqOffsetOrder(offset *OffsetOrder) Promise[OffsetOrder] {
	reqID, rtn := api.ReqOffsetOrder(offset)
	result := NewResult[OffsetOrder](rtn, maxDataLen)

	api.asyncCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspOffsetOrder(offset *OffsetOrder, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspOffsetOrder(offset, info, reqID, isLast)

	if async, exist := api.asyncCache.Load(reqID); exist {
		result := async.(*Result[OffsetOrder])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(offset, isLast)

		if isLast {
			api.asyncCache.Delete(reqID)
		}
	}
}

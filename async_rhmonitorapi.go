package rhmonitor4go

import (
	"bytes"
	"log"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	minDataLen = 1
	maxDataLen = 10
)

type RHRiskData interface {
	RspUserLogin | RspUserLogout | Investor | Account | Position | OffsetOrder | Order | Trade
}

type CallbackFn[T RHRiskData] func(*Result[T]) error

// Promise is a future result callback interface
// Then(), Catch(), Finally() can be called in any order sequence
//
// Await() must be last one in the call chain,
// in order to active an result watcher goroutine
// to execute callback functions
type Promise[T RHRiskData] interface {
	Await() error
	Then(CallbackFn[T]) Promise[T]
	Catch(CallbackFn[T]) Promise[T]
	Finally(CallbackFn[T]) Promise[T]
}

type PromiseError struct {
	ExecErr  error
	RspErr   error
	ThenErr  error
	CatchErr error
	FinalErr error
}

func (err *PromiseError) Error() string {
	buff := bytes.NewBuffer(nil)

	if err.ExecErr != nil {
		buff.WriteString("request execution failed: " + err.ExecErr.Error())
		return buff.String()
	}

	if err.RspErr == nil {
		if err.ThenErr != nil {
			buff.WriteString("Then() callback failed: " + err.ThenErr.Error() + "\n")
		}
	} else {
		buff.WriteString("response error: " + err.RspErr.Error() + "\n")

		if err.CatchErr != nil {
			buff.WriteString("Catch() callback failed: " + err.CatchErr.Error() + "\n")
		}
	}

	if err.FinalErr != nil {
		buff.WriteString("Finally() callback failed: " + err.FinalErr.Error())
	}

	return buff.String()
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

func (r *Result[T]) Await() error {
	r.finish.Add(1)

	r.startOnce.Do(func() { close(r.start) })

	var err PromiseError

	go func() {
		r.waitAsyncData()

		defer r.finish.Done()

		if r.ExecCode != 0 {
			if r.failFn != nil {
				err.ExecErr = r.failFn(r)
			}

			return
		}

		if r.RspInfo == nil || r.RspInfo.ErrorID == 0 {
			if r.successFn != nil {
				err.ThenErr = r.successFn(r)
			}
		} else if r.RspInfo.ErrorID != 0 {
			err.RspErr = r.RspInfo

			if r.failFn != nil {
				err.CatchErr = r.failFn(r)
			}
		}

		if r.finalFn != nil {
			err.FinalErr = r.finalFn(r)
		}
	}()

	r.finish.Wait()

	return &err
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
	orderFlow  *Result[Order]
	tradeFlow  *Result[Trade]
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

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorOrder(investor *Investor) Promise[Order] {
	_, rtn := api.ReqSubInvestorOrder(investor)

	atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&api.orderFlow)),
		nil, unsafe.Pointer(NewResult[Order](rtn, maxDataLen)),
	)

	return api.orderFlow
}

func (api *AsyncRHMonitorApi) OnRtnOrder(order *Order) {
	api.RHMonitorApi.OnRtnOrder(order)

	if api.orderFlow != nil {
		api.orderFlow.AppendResult(order, false)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqSubInvestorTrade(investor *Investor) Promise[Trade] {
	_, rtn := api.ReqSubInvestorTrade(investor)

	atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&api.tradeFlow)),
		nil, unsafe.Pointer(NewResult[Trade](rtn, maxDataLen)),
	)

	return api.tradeFlow
}

func (api *AsyncRHMonitorApi) OnRtnTrade(trade *Trade) {
	api.RHMonitorApi.OnRtnTrade(trade)

	if api.tradeFlow != nil {
		api.tradeFlow.AppendResult(trade, false)
	}
}

func (api *AsyncRHMonitorApi) OnRtnInvestorMoney(account *Account) {
	api.RHMonitorApi.OnRtnInvestorMoney(account)

	// api.accountChan.Publish("", *account)
}

func (api *AsyncRHMonitorApi) OnRtnInvestorPosition(position *Position) {
	api.RHMonitorApi.OnRtnInvestorPosition(position)

	// if api.positionFlow != nil {
	// 	api.positionFlow.AppendResult(position, false)
	// }
}

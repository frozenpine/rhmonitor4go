package rhmonitor4go

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	minDataLen = 1
	maxDataLen = 10
)

type RHRiskData interface {
	RspUserLogin | RspUserLogout | Investor | Account | Position | OffsetOrder | Order | Trade
}

type CallbackFn[T RHRiskData] func(*SingleResult[T]) error

// Promise is a future result callback interface
// Then(), Catch(), Finally() can be called in any order sequence
//
// Await() must be last one in the call chain,
// in order to active an result watcher goroutine
// to execute callback functions
type Promise[T RHRiskData] interface {
	Await(context.Context, time.Duration) error
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

type Result[T RHRiskData] interface {
	Promise[T]
	GetExecCode() int
	AppendResult(*T, bool)
}

type baseResult[T RHRiskData] struct {
	Data chan *T

	startFlag chan struct{}
	startOnce sync.Once

	notifyFlag chan struct{}
	notifyOnce sync.Once

	finish sync.WaitGroup

	successFn CallbackFn[T]
	failFn    CallbackFn[T]
	finalFn   CallbackFn[T]
}

func (r *baseResult[T]) waitAsyncData() {
	<-r.startFlag
	<-r.notifyFlag
}

func (r *baseResult[T]) Then(fn CallbackFn[T]) Promise[T] {
	r.successFn = fn

	return r
}

func (r *baseResult[T]) Catch(fn CallbackFn[T]) Promise[T] {
	r.failFn = fn

	return r
}

func (r *baseResult[T]) Finally(fn CallbackFn[T]) Promise[T] {
	r.finalFn = fn

	return r
}

func (r *baseResult[T]) Await(context.Context, time.Duration) error {
	return errors.New("await not implemented")
}

type BatchResult[T RHRiskData] struct {
	baseResult[T]

	RequestIDList []int
	ExecCodeList  []int
	RspInfoList   []*RspInfo
}

func (r *BatchResult[T]) AppendRequest(reqID, execCode int) {
	r.RequestIDList = append(r.RequestIDList, reqID)
	r.ExecCodeList = append(r.ExecCodeList, execCode)
}

func (r *BatchResult[T]) GetExecCode() int {
	size := len(r.RequestIDList)

	if size <= 0 {
		return 255
	}

	return r.ExecCodeList[size-1]
}

type SingleResult[T RHRiskData] struct {
	baseResult[T]

	RequestID int
	ExecCode  int
	RspInfo   *RspInfo
}

func (r *SingleResult[T]) GetExecCode() int {
	return r.ExecCode
}

func (r *SingleResult[T]) AppendResult(v *T, isLast bool) {
	r.notifyOnce.Do(func() { close(r.notifyFlag) })

	r.Data <- v

	if isLast {
		close(r.Data)
	}
}

func (r *SingleResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	r.finish.Add(1)

	r.startOnce.Do(func() { close(r.startFlag) })

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

func NewSingleResult[T RHRiskData](execCode int, dataLen int) *SingleResult[T] {
	if dataLen < minDataLen {
		dataLen = minDataLen
	}

	if dataLen > maxDataLen {
		dataLen = maxDataLen
	}

	result := SingleResult[T]{
		ExecCode: execCode,

		baseResult: baseResult[T]{
			startFlag:  make(chan struct{}),
			notifyFlag: make(chan struct{}),
			Data:       make(chan *T, dataLen),
		},
	}

	if execCode != 0 {
		close(result.notifyFlag)
		close(result.Data)
	}

	return &result
}

type AsyncRHMonitorApi struct {
	RHMonitorApi

	promiseCache sync.Map
	orderFlow    Result[Order]
	tradeFlow    Result[Trade]
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

	result := NewSingleResult[RspUserLogin](rtn, 1)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogin(login, info, reqID)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(*SingleResult[RspUserLogin])

		result.RspInfo = info

		result.AppendResult(login, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogout() Result[RspUserLogout] {
	reqID, rtn := api.ReqUserLogout()
	result := NewSingleResult[RspUserLogout](rtn, 1)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogout(logout, info, reqID)

	if async, exist := api.promiseCache.LoadAndDelete(reqID); exist {
		result := async.(*SingleResult[RspUserLogout])

		result.RspInfo = info

		result.AppendResult(logout, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryMonitorAccounts() Result[Investor] {
	reqID, rtn := api.ReqQryMonitorAccounts()
	result := NewSingleResult[Investor](rtn, maxDataLen)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryMonitorAccounts(investor, info, reqID, isLast)

	if promise, exist := api.promiseCache.Load(reqID); exist {
		result := promise.(*SingleResult[Investor])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(investor, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorMoney(investor *Investor) Result[Account] {
	reqID, rtn := api.ReqQryInvestorMoney(investor)
	result := NewSingleResult[Account](rtn, maxDataLen)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorMoney(acct *Account, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorMoney(acct, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(*SingleResult[Account])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(acct, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryInvestorPosition(investor *Investor, instrument string) Result[Position] {
	reqID, rtn := api.ReqQryInvestorPosition(investor, instrument)
	result := NewSingleResult[Position](rtn, maxDataLen)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorPosition(pos *Position, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorPosition(pos, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(*SingleResult[Position])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(pos, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

func (api *AsyncRHMonitorApi) AsyncReqOffsetOrder(offset *OffsetOrder) Result[OffsetOrder] {
	reqID, rtn := api.ReqOffsetOrder(offset)
	result := NewSingleResult[OffsetOrder](rtn, maxDataLen)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspOffsetOrder(offset *OffsetOrder, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspOffsetOrder(offset, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(reqID); exist {
		result := async.(*SingleResult[OffsetOrder])

		if result.RspInfo == nil {
			result.RspInfo = info
		}

		result.AppendResult(offset, isLast)

		if isLast {
			api.promiseCache.Delete(reqID)
		}
	}
}

// func (api *AsyncRHMonitorApi) AsyncReqSubInvestorOrder(investor *Investor) Result[Order] {
// 	_, rtn := api.ReqSubInvestorOrder(investor)

// 	atomic.CompareAndSwapPointer(
// 		(*unsafe.Pointer)(unsafe.Pointer(api.orderFlow)),
// 		nil, unsafe.Pointer(NewSingleResult[Order](rtn, maxDataLen)),
// 	)

// 	return api.orderFlow
// }

func (api *AsyncRHMonitorApi) OnRtnOrder(order *Order) {
	api.RHMonitorApi.OnRtnOrder(order)

	if api.orderFlow != nil {
		api.orderFlow.AppendResult(order, false)
	}
}

// func (api *AsyncRHMonitorApi) AsyncReqSubInvestorTrade(investor *Investor) *SingleResult[Trade] {
// 	_, rtn := api.ReqSubInvestorTrade(investor)

// 	atomic.CompareAndSwapPointer(
// 		(*unsafe.Pointer)(unsafe.Pointer(&api.tradeFlow)),
// 		nil, unsafe.Pointer(NewSingleResult[Trade](rtn, maxDataLen)),
// 	)

// 	return api.tradeFlow
// }

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

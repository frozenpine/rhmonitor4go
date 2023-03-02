package rhmonitor4go

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type RHRiskData interface {
	RspUserLogin | RspUserLogout | Investor | Account | Position | OffsetOrder | Order | Trade
}

type CallbackFn[T RHRiskData] func(Result[T]) error

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

type PromiseStage uint8

//go:generate stringer -type PromiseStage -linecomment
const (
	PromiseInfight PromiseStage = iota // Inflight
	PromiseAwait                       // Await
	PromiseThen                        // Then
	PromiseCatch                       // Catch
	PromiseFinal                       // Final
)

type ExecError struct {
	exec_code int
}

func (err *ExecError) Error() string {
	return fmt.Sprintf("reqeuest execution failed[%d]", err.exec_code)
}

type PromiseError struct {
	stage PromiseStage
	err   error
}

func (err *PromiseError) Error() string {
	return errors.Wrapf(
		err.err,
		"error occoured in stage %s",
		err.stage,
	).Error()
}

type Result[T RHRiskData] interface {
	Promise[T]

	IsBatch() bool
	GetRequestID() int64
	GetExecCode() int
	GetRspInfo() *RspInfo
	GetError() error
	GetData() <-chan *T

	AppendResult(int64, *T, bool)
}

type baseResult[T RHRiskData] struct {
	data chan *T

	startFlag chan struct{}
	startOnce sync.Once

	notifyFlag chan struct{}
	notifyOnce sync.Once

	finish sync.WaitGroup

	successFn CallbackFn[T]
	failFn    CallbackFn[T]
	finalFn   CallbackFn[T]
}

func (r *baseResult[T]) init() {
	r.data = make(chan *T)
	r.startFlag = make(chan struct{})
	r.notifyFlag = make(chan struct{})
}

func (r *baseResult[T]) GetData() <-chan *T {
	return r.data
}

func (r *baseResult[T]) waitAsyncData() {
	<-r.startFlag
	<-r.notifyFlag
}

func (r *baseResult[T]) IsBatch() bool { return false }

type BatchResult[T RHRiskData] struct {
	baseResult[T]

	requestIDList       []int64
	execCodeList        []int
	promiseErrChainList [][]PromiseError
	rspInfoCache        map[int64]chan *RspInfo
	rspStatusCache      map[int64]bool
}

func (r *BatchResult[T]) IsBatch() bool { return true }

func (r *BatchResult[T]) GetRequestID() int64 {
	size := len(r.requestIDList)
	if size == 0 {
		return -1
	}

	return r.requestIDList[size-1]
}

func (r *BatchResult[T]) GetRspInfo() *RspInfo {
	size := len(r.requestIDList)

	if size == 0 {
		return nil
	}

	// TODO: return RspInfo
	return nil
}

func (r *BatchResult[T]) AppendRequest(reqID int64, execCode int) {
	r.requestIDList = append(r.requestIDList, reqID)
	r.execCodeList = append(r.execCodeList, execCode)
	r.promiseErrChainList = append(r.promiseErrChainList, []PromiseError{})

	if execCode == 0 {
		r.rspInfoCache[reqID] = make(chan *RspInfo)
		r.rspStatusCache[reqID] = false
	}
}

func (r *BatchResult[T]) GetExecCode() int {
	size := len(r.requestIDList)

	if size <= 0 {
		return 255
	}

	return r.execCodeList[size-1]
}

func (r *BatchResult[T]) GetError() (err error) {
	for _, errChain := range r.promiseErrChainList {
		for _, stageErr := range errChain {
			if err == nil {
				err = &stageErr
			} else {
				err = errors.Wrap(err, stageErr.Error())
			}
		}
	}

	return
}

func (r *BatchResult[T]) AppendResult(reqID int64, v *T, isLast bool) {
	if _, exist := r.rspInfoCache[reqID]; !exist {
		log.Printf(
			"Appended RequestID[%d] not exist in BatchResult", reqID,
		)

		return
	}

	r.notifyOnce.Do(func() { close(r.notifyFlag) })

	r.data <- v

	r.rspStatusCache[reqID] = isLast

	all := true
	for _, v := range r.rspStatusCache {
		all = all && v
	}

	if all {
		close(r.data)
	}
}

func (r *BatchResult[T]) Then(fn CallbackFn[T]) Promise[T] {
	r.successFn = fn

	return r
}

func (r *BatchResult[T]) Catch(fn CallbackFn[T]) Promise[T] {
	r.failFn = fn

	return r
}

func (r *BatchResult[T]) Finally(fn CallbackFn[T]) Promise[T] {
	r.finalFn = fn

	return r
}

func (r *BatchResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	r.startOnce.Do(func() { close(r.startFlag) })

	r.finish.Add(len(r.rspInfoCache))

	for idx, reqID := range r.requestIDList {
		if r.execCodeList[idx] != 0 {
			continue
		}

		go func(idx int, reqID int64) {
			r.waitAsyncData()
			defer r.finish.Done()

			execCode := r.execCodeList[idx]
			result := NewSingleResult[T](reqID, execCode)

			if execCode != 0 {
				r.promiseErrChainList[idx] = append(
					r.promiseErrChainList[idx],
					PromiseError{
						stage: PromiseInfight,
						err:   &ExecError{exec_code: execCode},
					},
				)

				goto CATCH
			}

			if ch, exist := r.rspInfoCache[reqID]; exist {
				result.rspInfo = <-ch
			}

			if result.rspInfo == nil || result.rspInfo.ErrorID == 0 {
				goto THEN
			} else {
				r.promiseErrChainList[idx] = append(
					r.promiseErrChainList[idx],
					PromiseError{
						stage: PromiseAwait,
						err:   result.rspInfo,
					},
				)

				goto CATCH
			}

		THEN:
			if r.successFn != nil {
				if err := r.successFn(result); err != nil {
					r.promiseErrChainList[idx] = append(
						r.promiseErrChainList[idx],
						PromiseError{
							stage: PromiseThen,
							err:   err,
						},
					)

					goto CATCH
				}
			}

			goto FINAL
		CATCH:
			if r.failFn != nil {
				if err := r.failFn(result); err != nil {
					r.promiseErrChainList[idx] = append(
						r.promiseErrChainList[idx],
						PromiseError{
							stage: PromiseCatch,
							err:   err,
						},
					)
				}
			}

			goto FINAL
		FINAL:
			if r.finalFn != nil {
				if err := r.finalFn(result); err != nil {
					r.promiseErrChainList[idx] = append(
						r.promiseErrChainList[idx],
						PromiseError{
							stage: PromiseFinal,
							err:   err,
						},
					)
				}
			}
		}(idx, reqID)
	}

	r.finish.Wait()

	return r.GetError()
}

func NewBatchResult[T RHRiskData]() *BatchResult[T] {
	result := BatchResult[T]{
		rspInfoCache:   make(map[int64]chan *RspInfo),
		rspStatusCache: make(map[int64]bool),
	}
	result.init()

	return &result
}

type SingleResult[T RHRiskData] struct {
	promiseErrChain []PromiseError
	requestID       int64
	execCode        int
	rspInfo         *RspInfo

	baseResult[T]
}

func (r *SingleResult[T]) GetRequestID() int64 {
	return r.requestID
}

func (r *SingleResult[T]) GetRspInfo() *RspInfo {
	return r.rspInfo
}

func (r *SingleResult[T]) GetExecCode() int {
	return r.execCode
}

func (r *SingleResult[T]) GetError() (err error) {
	for _, e := range r.promiseErrChain {
		if err == nil {
			err = &e
		} else {
			err = errors.Wrap(err, e.Error())
		}
	}

	return err
}

func (r *SingleResult[T]) AppendResult(reqID int64, v *T, isLast bool) {
	if reqID != r.requestID {
		log.Printf(
			"Appended RequestID[%d] miss match with Result[%d]",
			reqID, r.requestID,
		)

		return
	}

	r.notifyOnce.Do(func() { close(r.notifyFlag) })

	r.data <- v

	if isLast {
		close(r.data)
	}
}

func (r *SingleResult[T]) Then(fn CallbackFn[T]) Promise[T] {
	r.successFn = fn

	return r
}

func (r *SingleResult[T]) Catch(fn CallbackFn[T]) Promise[T] {
	r.failFn = fn

	return r
}

func (r *SingleResult[T]) Finally(fn CallbackFn[T]) Promise[T] {
	r.finalFn = fn

	return r
}

func (r *SingleResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	r.finish.Add(1)

	r.startOnce.Do(func() { close(r.startFlag) })

	go func() {
		r.waitAsyncData()

		defer r.finish.Done()

		if r.execCode != 0 {
			r.promiseErrChain = append(
				r.promiseErrChain,
				PromiseError{
					stage: PromiseInfight,
					err:   &ExecError{exec_code: r.execCode},
				},
			)

			goto CATCH
		}

		if r.rspInfo == nil || r.rspInfo.ErrorID == 0 {
			goto THEN
		} else {
			r.promiseErrChain = append(
				r.promiseErrChain,
				PromiseError{
					stage: PromiseAwait,
					err:   r.rspInfo,
				},
			)
			goto CATCH
		}

	THEN:
		if r.successFn != nil {
			if err := r.successFn(r); err != nil {
				r.promiseErrChain = append(
					r.promiseErrChain,
					PromiseError{
						stage: PromiseThen,
						err:   err,
					},
				)

				goto CATCH
			}
		}

		goto FINAL

	CATCH:
		if r.failFn != nil {
			if err := r.failFn(r); err != nil {
				r.promiseErrChain = append(
					r.promiseErrChain,
					PromiseError{
						stage: PromiseCatch,
						err:   err,
					},
				)
			}
		}

		goto FINAL

	FINAL:
		if r.finalFn != nil {
			if err := r.finalFn(r); err != nil {
				r.promiseErrChain = append(
					r.promiseErrChain,
					PromiseError{
						stage: PromiseFinal,
						err:   err,
					},
				)
			}
		}
	}()

	r.finish.Wait()

	return r.GetError()
}

func NewSingleResult[T RHRiskData](reqID int64, execCode int) *SingleResult[T] {
	result := SingleResult[T]{
		requestID: reqID,
		execCode:  execCode,
	}

	result.init()

	if execCode != 0 {
		close(result.notifyFlag)
		close(result.data)
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

	result := NewSingleResult[RspUserLogin](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogin(login *RspUserLogin, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogin(login, info, reqID)

	if async, exist := api.promiseCache.LoadAndDelete(int64(reqID)); exist {
		result := async.(*SingleResult[RspUserLogin])

		result.rspInfo = info

		result.AppendResult(int64(reqID), login, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqUserLogout() Result[RspUserLogout] {
	reqID, rtn := api.ReqUserLogout()
	result := NewSingleResult[RspUserLogout](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspUserLogout(logout *RspUserLogout, info *RspInfo, reqID int) {
	api.RHMonitorApi.OnRspUserLogout(logout, info, reqID)

	if async, exist := api.promiseCache.LoadAndDelete(int64(reqID)); exist {
		result := async.(*SingleResult[RspUserLogout])

		result.rspInfo = info

		result.AppendResult(int64(reqID), logout, true)
	}
}

func (api *AsyncRHMonitorApi) AsyncReqQryMonitorAccounts() Result[Investor] {
	reqID, rtn := api.ReqQryMonitorAccounts()

	result := NewSingleResult[Investor](reqID, rtn)

	api.promiseCache.Store(reqID, result)

	return result
}

func (api *AsyncRHMonitorApi) OnRspQryMonitorAccounts(investor *Investor, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryMonitorAccounts(investor, info, reqID, isLast)

	if promise, exist := api.promiseCache.Load(int64(reqID)); exist {
		result := promise.(*SingleResult[Investor])

		if result.rspInfo == nil {
			result.rspInfo = info
		}

		result.AppendResult(int64(reqID), investor, isLast)

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

		results.AppendRequest(reqID, rtn)

		if rtn != 0 {
			log.Printf("Query money for user[%s] failed: %d", i.InvestorID, rtn)
			return false
		}

		return true
	})

	return results
}

func (api *AsyncRHMonitorApi) OnRspQryInvestorMoney(acct *Account, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorMoney(acct, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(int64(reqID)); exist {
		result := async.(*SingleResult[Account])

		if result.rspInfo == nil {
			result.rspInfo = info
		}

		result.AppendResult(int64(reqID), acct, isLast)

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

func (api *AsyncRHMonitorApi) OnRspQryInvestorPosition(pos *Position, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspQryInvestorPosition(pos, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(int64(reqID)); exist {
		result := async.(*SingleResult[Position])

		if result.rspInfo == nil {
			result.rspInfo = info
		}

		result.AppendResult(int64(reqID), pos, isLast)

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

func (api *AsyncRHMonitorApi) OnRspOffsetOrder(offset *OffsetOrder, info *RspInfo, reqID int, isLast bool) {
	api.RHMonitorApi.OnRspOffsetOrder(offset, info, reqID, isLast)

	if async, exist := api.promiseCache.Load(int64(reqID)); exist {
		result := async.(*SingleResult[OffsetOrder])

		if result.rspInfo == nil {
			result.rspInfo = info
		}

		result.AppendResult(int64(reqID), offset, isLast)

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
		api.orderFlow.AppendResult(-1, order, false)
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
		api.tradeFlow.AppendResult(-1, trade, false)
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

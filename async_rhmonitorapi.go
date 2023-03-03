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

const (
	InfinitResultReq int64 = -1
	orderFlow              = "order"
	tradeFlow              = "trade"
	accountFlow            = "account"
	positionFlow           = "position"
)

var (
	ErrRspChanMissing = errors.New("response chan missing")
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
	SetRspInfo(int64, *RspInfo)
}

type baseResult[T RHRiskData] struct {
	self Result[T]

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

func (r *baseResult[T]) init(self Result[T]) {
	r.self = self
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

func (r *baseResult[T]) Then(fn CallbackFn[T]) Promise[T] {
	if r.self == nil {
		panic("Self pointer missing")
	}

	r.successFn = fn

	return r.self
}

func (r *baseResult[T]) Catch(fn CallbackFn[T]) Promise[T] {
	if r.self == nil {
		panic("Self pointer missing")
	}

	r.failFn = fn

	return r.self
}

func (r *baseResult[T]) Finally(fn CallbackFn[T]) Promise[T] {
	if r.self == nil {
		panic("Self pointer missing")
	}

	r.finalFn = fn

	return r.self
}

type BatchResult[T RHRiskData] struct {
	baseResult[T]

	requestIDList        []int64
	execCodeList         []int
	promiseErrChainList  [][]PromiseError
	rspInfoCache         map[int64]chan *RspInfo
	rspNotifyStatusCache map[int64]bool
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
		r.rspNotifyStatusCache[reqID] = false
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

func (r *BatchResult[T]) SetRspInfo(reqID int64, rsp *RspInfo) {
	if rsp == nil {
		return
	}

	if rspCh, exist := r.rspInfoCache[reqID]; exist {
		rspCh <- rsp
	}

	r.rspNotifyStatusCache[reqID] = true
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

	r.rspNotifyStatusCache[reqID] = isLast

	all := true
	for _, v := range r.rspNotifyStatusCache {
		all = all && v
	}

	if all {
		close(r.data)
	}
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
		rspInfoCache:         make(map[int64]chan *RspInfo),
		rspNotifyStatusCache: make(map[int64]bool),
	}
	result.init(&result)

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

func (r *SingleResult[T]) SetRspInfo(_ int64, rsp *RspInfo) {
	r.rspInfo = rsp
}

func (r *SingleResult[T]) awaitLoop() {
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
}

func (r *SingleResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	r.finish.Add(1)

	r.startOnce.Do(func() { close(r.startFlag) })

	go r.awaitLoop()

	r.finish.Wait()

	return r.GetError()
}

func NewSingleResult[T RHRiskData](reqID int64, execCode int) *SingleResult[T] {
	result := SingleResult[T]{
		requestID: reqID,
		execCode:  execCode,
	}

	result.init(&result)

	if execCode != 0 {
		close(result.notifyFlag)
		close(result.data)
	}

	return &result
}

type FlowResult[T RHRiskData] struct {
	SingleResult[T]
}

func (r *FlowResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	go r.awaitLoop()

	if r.GetExecCode() != 0 {
		r.finish.Wait()
	}

	return r.GetError()
}

func NewFlowResult[T RHRiskData](execCode int) *FlowResult[T] {
	result := FlowResult[T]{
		SingleResult[T]{
			requestID: InfinitResultReq,
			execCode:  execCode,
		},
	}
	result.init(&result)

	if execCode != 0 {
		close(result.notifyFlag)
		close(result.data)
	}

	return &result
}

type AsyncRHMonitorApi struct {
	RHMonitorApi

	promiseCache sync.Map
	flowCache    sync.Map
	// orderFlow    Result[Order]
	// tradeFlow    Result[Trade]
	// accountFlow  Result[Account]
	// positionFlow Result[Position]
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
	api.RHMonitorApi.OnRspQryInvestorMoney(acct, info, reqID, isLast)

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
	// api.RHMonitorApi.OnRspQryInvestorPosition(pos, info, reqID, isLast)

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
	// api.RHMonitorApi.OnRtnOrder(order)

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

package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/frozenpine/rhmonitor4go"
	"github.com/pkg/errors"
)

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
	GetRspInfo() *rhmonitor4go.RspInfo
	GetError() error
	GetData() <-chan *T

	AppendResult(int64, *T, bool)
	SetRspInfo(int64, *rhmonitor4go.RspInfo)
}

type baseResult[T RHRiskData] struct {
	self Result[T]

	data chan *T

	rspFin sync.WaitGroup

	successFn CallbackFn[T]
	failFn    CallbackFn[T]
	finalFn   CallbackFn[T]
}

func (r *baseResult[T]) init(self Result[T]) {
	r.self = self
	r.data = make(chan *T, 1)
}

func (r *baseResult[T]) GetData() <-chan *T {
	return r.data
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

type rsp[T RHRiskData] struct {
	info       *rhmonitor4go.RspInfo
	status     bool
	dataFlag   chan struct{}
	notifyData sync.Once
	data       chan *T
}

type BatchResult[T RHRiskData] struct {
	baseResult[T]

	requestIDList       []int64
	execCodeList        []int
	promiseErrChainList [][]PromiseError
	rspCache            map[int64]*rsp[T]

	errCollector  chan int
	dataCollector chan *T
	collectorDone chan struct{}
}

func (r *BatchResult[T]) waitRsp(reqID int64) error {
	cache, exist := r.rspCache[reqID]

	if !exist {
		return ErrRspChanMissing
	}

	<-cache.dataFlag

	return nil
}

func (r *BatchResult[T]) IsBatch() bool { return true }

func (r *BatchResult[T]) GetRequestID() int64 {
	size := len(r.requestIDList)
	if size == 0 {
		return -1
	}

	return r.requestIDList[size-1]
}

func (r *BatchResult[T]) GetRspInfo() *rhmonitor4go.RspInfo {
	size := len(r.requestIDList)

	if size == 0 {
		return nil
	}

	var rsp *rhmonitor4go.RspInfo

	for _, cache := range r.rspCache {
		if cache.info != nil && cache.info.ErrorID != 0 {
			rsp = cache.info
		}
	}

	return rsp
}

func (r *BatchResult[T]) AppendRequest(reqID int64, execCode int) {
	r.requestIDList = append(r.requestIDList, reqID)
	r.execCodeList = append(r.execCodeList, execCode)
	r.promiseErrChainList = append(r.promiseErrChainList, []PromiseError{})

	if execCode == 0 {
		r.rspCache[reqID] = &rsp[T]{
			dataFlag: make(chan struct{}),
			data:     make(chan *T, 1),
		}
	}

	logger.Printf("Append request: %d, %d", reqID, execCode)
}

func (r *BatchResult[T]) GetExecCode() int {
	rtn := 0

	for _, code := range r.execCodeList {
		rtn += code
	}

	return rtn
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

func (r *BatchResult[T]) SetRspInfo(reqID int64, rsp *rhmonitor4go.RspInfo) {
	if rsp == nil {
		return
	}

	if cache, exist := r.rspCache[reqID]; exist && cache.info == nil {
		cache.info = rsp
	}
}

func (r *BatchResult[T]) AppendResult(reqID int64, v *T, isLast bool) {
	cache, exist := r.rspCache[reqID]

	if !exist {
		logger.Printf(
			"Appended RequestID[%d] not exist in BatchResult", reqID,
		)

		return
	}

	logger.Printf("Append result[%d] %v: %+v", reqID, isLast, v)

	cache.status = isLast

	cache.data <- v

	cache.notifyData.Do(func() { close(cache.dataFlag) })

	if isLast {
		close(cache.data)
	}
}

func (r *BatchResult[T]) Await(ctx context.Context, timeout time.Duration) error {
	logger.Printf("Await response for: %v", r.requestIDList)

	r.rspFin.Add(len(r.requestIDList))

	go func() {
		defer func() { close(r.collectorDone) }()

		// TODO: 大改造THEN, CATCH, FINAL
	}()

	for idx, reqID := range r.requestIDList {
		go func(idx int, reqID int64) {
			if err := r.waitRsp(reqID); err != nil {
				logger.Printf("Wait response[%d] faield: %+v", reqID, err)
				r.rspFin.Done()
				return
			}

			defer r.rspFin.Done()

			execCode := r.execCodeList[idx]
			cache := r.rspCache[reqID]

			if execCode != 0 {
				r.promiseErrChainList[idx] = append(
					r.promiseErrChainList[idx],
					PromiseError{
						stage: PromiseInfight,
						err:   &ExecError{exec_code: execCode},
					},
				)

				return
			}

			if cache.info == nil || cache.info.ErrorID == 0 {
				for v := range cache.data {
					r.dataCollector <- v
				}
			} else {
				r.promiseErrChainList[idx] = append(
					r.promiseErrChainList[idx],
					PromiseError{
						stage: PromiseAwait,
						err:   cache.info,
					},
				)

				// return
			}

			// THEN:
			// 	if r.successFn != nil {
			// 		if err := r.successFn(r); err != nil {
			// 			r.promiseErrChainList[idx] = append(
			// 				r.promiseErrChainList[idx],
			// 				PromiseError{
			// 					stage: PromiseThen,
			// 					err:   err,
			// 				},
			// 			)

			// 			goto CATCH
			// 		}
			// 	}

			// 	goto FINAL
			// CATCH:
			// 	if r.failFn != nil {
			// 		if err := r.failFn(r); err != nil {
			// 			r.promiseErrChainList[idx] = append(
			// 				r.promiseErrChainList[idx],
			// 				PromiseError{
			// 					stage: PromiseCatch,
			// 					err:   err,
			// 				},
			// 			)
			// 		}
			// 	}

			// 	goto FINAL
			// FINAL:
			// 	if r.finalFn != nil {
			// 		if err := r.finalFn(r); err != nil {
			// 			r.promiseErrChainList[idx] = append(
			// 				r.promiseErrChainList[idx],
			// 				PromiseError{
			// 					stage: PromiseFinal,
			// 					err:   err,
			// 				},
			// 			)
			// 		}
			// 	}
		}(idx, reqID)
	}

	r.rspFin.Wait()
	close(r.data)

	<-r.collectorDone
	close(r.data)

	return r.GetError()
}

func NewBatchResult[T RHRiskData]() *BatchResult[T] {
	result := BatchResult[T]{
		rspCache:      make(map[int64]*rsp[T]),
		errCollector:  make(chan int, 1),
		dataCollector: make(chan *T, 1),
		collectorDone: make(chan struct{}),
	}
	result.init(&result)

	return &result
}

type SingleResult[T RHRiskData] struct {
	baseResult[T]

	promiseErrChain []PromiseError
	requestID       int64
	execCode        int
	rspInfo         *rhmonitor4go.RspInfo
	notifyFlag      chan struct{}
	notifyOnce      sync.Once
}

func (r *SingleResult[T]) waitRsp() {
	<-r.notifyFlag
}

func (r *SingleResult[T]) GetRequestID() int64 {
	return r.requestID
}

func (r *SingleResult[T]) GetRspInfo() *rhmonitor4go.RspInfo {
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
		logger.Printf(
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

func (r *SingleResult[T]) SetRspInfo(_ int64, rsp *rhmonitor4go.RspInfo) {
	r.rspInfo = rsp
}

func (r *SingleResult[T]) awaitLoop() {
	r.waitRsp()

	defer r.rspFin.Done()

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
	r.rspFin.Add(1)

	go r.awaitLoop()

	r.rspFin.Wait()

	return r.GetError()
}

func NewSingleResult[T RHRiskData](reqID int64, execCode int) *SingleResult[T] {
	result := SingleResult[T]{
		requestID:  reqID,
		execCode:   execCode,
		notifyFlag: make(chan struct{}),
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
	r.rspFin.Add(1)

	go r.awaitLoop()

	if r.GetExecCode() != 0 {
		r.rspFin.Wait()
	}

	return r.GetError()
}

func NewFlowResult[T RHRiskData](execCode int) *FlowResult[T] {
	result := FlowResult[T]{
		SingleResult[T]{
			requestID:  InfinitResultReq,
			execCode:   execCode,
			notifyFlag: make(chan struct{}),
		},
	}
	result.init(&result)

	if execCode != 0 {
		close(result.notifyFlag)
		close(result.data)
	}

	return &result
}

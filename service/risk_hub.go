package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	rohon "github.com/frozenpine/rhmonitor4go"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrInvalidArgs     = errors.New("invalid request args")
	ErrRiskApiNotFound = errors.New("risk api not found")
	ErrRequestFailed   = errors.New("request execution failed")
)

const broadcastBufferSize = 10

type grpcRiskApi struct {
	rohon.AsyncRHMonitorApi

	ctx             context.Context
	front           *RiskServer
	broadcastSub    atomic.Bool
	broadcastCh     chan string
	broadcastLock   sync.RWMutex
	broadcastBuffer []string
}

func (api *grpcRiskApi) sendBroadcast(msg string) {
	api.broadcastCh <- msg
	// if api.broadcastSub.Load() {
	// 	api.broadcastCh <- msg
	// 	return
	// }

	// for {
	// 	if api.broadcastLock.TryLock() {
	// 		if len(api.broadcastBuffer) >= broadcastBufferSize {
	// 			api.broadcastBuffer = api.broadcastBuffer[broadcastBufferSize/2:]
	// 		}

	// 		api.broadcastBuffer = append(api.broadcastBuffer, msg)

	// 		api.broadcastLock.Unlock()
	// 	}
	// }
}

func (api *grpcRiskApi) OnFrontConnected() {
	api.HandleConnected()

	api.sendBroadcast(fmt.Sprintf(
		"Front[%s:%d] connected",
		api.front.ServerAddr, api.front.ServerPort,
	))
}

func (api *grpcRiskApi) OnFrontDisconnected(reason rohon.Reason) {
	api.HandleDisconnected()

	api.sendBroadcast(fmt.Sprintf(
		"Front[%s:%d] disconnected: %v",
		api.front.ServerAddr, api.front.ServerPort, reason,
	))
}

func reqFinalFn[T rohon.RHRiskData](result *Result) rohon.CallbackFn[T] {
	return func(req rohon.Result[T]) error {
		result.ReqId = int32(req.GetRequestID())

		if rsp := req.GetRspInfo(); rsp != nil {
			result.RspInfo = &RspInfo{
				ErrorId:  int32(rsp.ErrorID),
				ErrorMsg: rsp.ErrorMsg,
			}
		}
		return nil
	}
}

func checkPromise[T rohon.RHRiskData](result rohon.Result[T], caller string) (rohon.Promise[T], error) {
	return result, result.GetError()
}

type RiskHub struct {
	UnimplementedRohonMonitorServer

	svr           *grpc.Server
	apiCache      sync.Map
	apiReqTimeout time.Duration
}

func (hub *RiskHub) getApiInstance(idt string) (*grpcRiskApi, error) {
	if api, exist := hub.apiCache.Load(idt); exist {
		return api.(*grpcRiskApi), nil
	}

	return nil, errors.Wrapf(
		ErrRiskApiNotFound, "invalid risk api identity: %s",
		idt,
	)
}

func (hub *RiskHub) Init(ctx context.Context, req *Request) (result *Result, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	front := req.GetFront()
	if front == nil {
		err = errors.Wrap(ErrInvalidArgs, "[Init] func should set [RiskServer] arg")
		return
	}

	api := grpcRiskApi{
		broadcastCh:     make(chan string, broadcastBufferSize),
		broadcastBuffer: make([]string, 0, broadcastBufferSize),
		ctx:             ctx,
		front:           front,
	}

	if err = api.Init(
		front.BrokerId, front.ServerAddr,
		int(front.ServerPort), &api,
	); err != nil {
		log.Printf("Create AsyncRHMonitorApi failed: %+v", err)

		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		err = errors.Wrap(err, "make risk api identity failed")
		return
	}

	identity := id.String()

	result = &Result{}

	hub.apiCache.Store(identity, &api)
	result.ReqId = -1
	result.Response = &Result_ApiIdentity{ApiIdentity: identity}

	log.Printf("New risk api created: %s", id)

	return
}

func (hub *RiskHub) Release(ctx context.Context, req *Request) (empty *emptypb.Empty, err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	api.Release()

	empty = &emptypb.Empty{}

	log.Printf("Releasing risk api: %s", req.GetApiIdentity())

	return
}

func (hub *RiskHub) ReqUserLogin(ctx context.Context, req *Request) (result *Result, err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.RspUserLogin]

	if promise, err = checkPromise(
		api.AsyncReqUserLogin(&rohon.RiskUser{
			UserID:   req.GetLogin().UserId,
			Password: req.GetLogin().Password,
		}),
		"ReqUserLogin",
	); err != nil {
		return
	}

	result = &Result{}

	if err = promise.Then(func(r rohon.Result[rohon.RspUserLogin]) error {
		login := <-r.GetData()

		result.Response = &Result_UserLogin{
			UserLogin: convertRspLogin(login),
		}

		return nil
	}).Finally(
		reqFinalFn[rohon.RspUserLogin](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqUserLogin] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) ReqUserLogout(ctx context.Context, req *Request) (result *Result, err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.RspUserLogout]
	if promise, err = checkPromise(
		api.AsyncReqUserLogout(),
		"ReqUserLogout",
	); err != nil {
		return
	}

	result = &Result{}

	if err = promise.Then(func(r rohon.Result[rohon.RspUserLogout]) error {
		logout := <-r.GetData()

		result.Response = &Result_UserLogout{
			UserLogout: convertRspLogout(logout),
		}

		return nil
	}).Finally(
		reqFinalFn[rohon.RspUserLogout](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqUserLogout] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) ReqQryMonitorAccounts(ctx context.Context, req *Request) (result *Result, err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.Investor]
	if promise, err = checkPromise(
		api.AsyncReqQryMonitorAccounts(),
		"ReqQryMonitorAccounts",
	); err != nil {
		return
	}

	result = &Result{}

	if err = promise.Then(func(r rohon.Result[rohon.Investor]) error {
		investors := InvestorList{}

		for inv := range r.GetData() {
			investors.Data = append(investors.Data, convertInvestor(inv))
		}

		result.Response = &Result_Investors{
			Investors: &investors,
		}

		return nil
	}).Finally(
		reqFinalFn[rohon.Investor](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqQryMonitorAccounts] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) ReqQryInvestorMoney(ctx context.Context, req *Request) (result *Result, err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.Account]

	if inv := req.GetInvestor(); inv != nil {
		investor := &rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		}

		if promise, err = checkPromise(
			api.AsyncReqQryInvestorMoney(investor),
			"ReqQryInvestorMoney",
		); err != nil {
			return
		}
	} else {
		if promise, err = checkPromise(
			api.AsyncReqQryAllInvestorMoney(),
			"ReqQryInvestorMoney",
		); err != nil {
			return
		}
	}

	result = &Result{}

	if err = promise.Then(func(r rohon.Result[rohon.Account]) error {
		accounts := &AccountList{}

		for acct := range r.GetData() {
			accounts.Data = append(accounts.Data, convertAccount(acct))
		}

		result.Response = &Result_Accounts{Accounts: accounts}

		return nil
	}).Finally(
		reqFinalFn[rohon.Account](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqQryInvestorMoney] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) SubInvestorOrder(req *Request, stream RohonMonitor_SubInvestorOrderServer) (err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Order]

	if inv := req.GetInvestor(); inv != nil {
		result = api.AsyncReqSubInvestorOrder(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		})
	} else {
		result = api.AsyncReqSubAllInvestorOrder()
	}

	if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "sub investor's order failed")

		return
	}

	for ord := range result.GetData() {
		stream.Send(convertOrder(ord))
	}

	return
}

func (hub *RiskHub) SubInvestorTrade(req *Request, stream RohonMonitor_SubInvestorTradeServer) (err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Trade]

	if inv := req.GetInvestor(); inv != nil {
		result = api.AsyncReqSubInvestorTrade(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		})
	} else {
		result = api.AsyncReqSubAllInvestorTrade()
	}

	if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "sub investor's trade failed")

		return
	}

	for td := range result.GetData() {
		stream.Send(convertTrade(td))
	}

	return
}

func (hub *RiskHub) SubInvestorMoney(req *Request, stream RohonMonitor_SubInvestorMoneyServer) (err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	filter := req.GetInvestor()

	result := api.AsyncReqSubAllInvestorMoney()

	if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "sub investor's money failed")

		return
	}

	for acct := range result.GetData() {
		if filter != nil && filter.InvestorId != acct.AccountID {
			continue
		}

		stream.Send(convertAccount(acct))
	}

	return
}

func (hub *RiskHub) SubInvestorPosition(req *Request, stream RohonMonitor_SubInvestorPositionServer) (err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Position]

	if inv := req.GetInvestor(); inv != nil {
		result = api.AsyncReqQryInvestorPosition(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		}, "")
	} else {
		result = api.AsyncReqSubAllInvestorPosition()
	}

	if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "sub investor's postion failed")

		return
	}

	for pos := range result.GetData() {
		stream.Send(convertPosition(pos))
	}

	return
}

func (hub *RiskHub) SubBroadcast(req *Request, stream RohonMonitor_SubBroadcastServer) (err error) {
	var api *grpcRiskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	// for {
	// 	if api.broadcastLock.TryRLock() {
	// 		his := api.broadcastBuffer

	// 		api.broadcastLock.Unlock()

	// 		for _, msg := range his {
	// 			stream.Send(&Broadcast{Message: msg})
	// 		}

	// 		break
	// 	}
	// }

	for msg := range api.broadcastCh {
		stream.Send(&Broadcast{Message: msg})
	}

	return
}

func NewRohonMonitorHub(svr *grpc.Server) RohonMonitorServer {
	pb := &RiskHub{
		svr: svr,
	}

	RegisterRohonMonitorServer(svr, pb)

	// reflection.Register(svr)

	return pb
}

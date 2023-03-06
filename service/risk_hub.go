package service

import (
	"context"
	"sync"
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

type riskApi struct {
	ins *rohon.AsyncRHMonitorApi

	ctx   context.Context
	front *RiskServer
}

// type apiPool struct {
// 	availablePool sync.Map
// 	usedPool      sync.Map
// 	timeout       time.Duration
// }

// func (pool *apiPool) createApiInstance(ctx context.Context, front RiskServer) (*riskApi, error) {
// 	api, exist := pool.availablePool.Load(front)

// 	if !exist {
// 		ins := rohon.AsyncRHMonitorApi{}
// 		if err := ins.Init(
// 			front.BrokerId, front.ServerAddr,
// 			int(front.ServerPort), &ins,
// 		); err != nil {
// 			return nil, err
// 		} else {
// 			api = &riskApi{
// 				ins:   &ins,
// 				ctx:   ctx,
// 				front: &front,
// 			}
// 		}
// 	}

// 	return api.(*riskApi), nil
// }

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

func (hub *RiskHub) getApiInstance(idt string) (*riskApi, error) {
	if api, exist := hub.apiCache.Load(idt); exist {
		return api.(*riskApi), nil
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

	asyncInstance := rohon.NewAsyncRHMonitorApi(front.BrokerId, front.ServerAddr, int(front.ServerPort))
	if asyncInstance == nil {
		err = errors.Wrap(err, "risk api init failed")
		return
	}

	api := riskApi{
		ins:   asyncInstance,
		ctx:   ctx,
		front: front,
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

	return
}

func (hub *RiskHub) Release(ctx context.Context, req *Request) (empty *emptypb.Empty, err error) {
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	api.ins.Release()

	empty = &emptypb.Empty{}

	return
}

func (hub *RiskHub) ReqUserLogin(ctx context.Context, req *Request) (result *Result, err error) {
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.RspUserLogin]

	if promise, err = checkPromise(
		api.ins.AsyncReqUserLogin(&rohon.RiskUser{
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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.RspUserLogout]
	if promise, err = checkPromise(
		api.ins.AsyncReqUserLogout(),
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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var promise rohon.Promise[rohon.Investor]
	if promise, err = checkPromise(
		api.ins.AsyncReqQryMonitorAccounts(),
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
	var api *riskApi
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
			api.ins.AsyncReqQryInvestorMoney(investor),
			"ReqQryInvestorMoney",
		); err != nil {
			return
		}
	} else {
		if promise, err = checkPromise(
			api.ins.AsyncReqQryAllInvestorMoney(),
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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Order]

	if inv := req.GetInvestor(); inv != nil {
		result = api.ins.AsyncReqSubInvestorOrder(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		})
	} else {
		result = api.ins.AsyncReqSubAllInvestorOrder()
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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Trade]

	if inv := req.GetInvestor(); inv != nil {
		result = api.ins.AsyncReqSubInvestorTrade(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		})
	} else {
		result = api.ins.AsyncReqSubAllInvestorTrade()
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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	filter := req.GetInvestor()

	result := api.ins.AsyncReqSubAllInvestorMoney()

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
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var result rohon.Result[rohon.Position]

	if inv := req.GetInvestor(); inv != nil {
		result = api.ins.AsyncReqQryInvestorPosition(&rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		}, "")
	} else {
		result = api.ins.AsyncReqSubAllInvestorPosition()
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

func NewRohonMonitorHub(svr *grpc.Server) RohonMonitorServer {
	pb := &RiskHub{
		svr: svr,
	}

	RegisterRohonMonitorServer(svr, pb)

	// reflection.Register(svr)

	return pb
}

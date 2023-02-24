package service

import (
	"context"
	origin_err "errors"
	"sync"
	"time"

	rohon "github.com/frozenpine/rhmonitor4go"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	ErrInvalidArgs     = origin_err.New("invalid request args")
	ErrRiskApiNotFound = origin_err.New("risk api not found")
	ErrRequestFailed   = origin_err.New("request execution failed")
)

type riskApi struct {
	ins *rohon.AsyncRHMonitorApi

	ctx   context.Context
	front *RiskServer
}

func reqFinalFn[T rohon.RHRiskData](result *Result) rohon.CallbackFn[T] {
	return func(req *rohon.Result[T]) error {
		result.ReqId = int32(req.RequestID)

		if req.RspInfo != nil {
			result.RspInfo = &RspInfo{
				ErrorId:  int32(req.RspInfo.ErrorID),
				ErrorMsg: req.RspInfo.ErrorMsg,
			}
		}
		return nil
	}
}

func checkFuture[T rohon.RHRiskData](future *rohon.Result[T], caller string) (rohon.Promise[T], error) {
	if future.ExecCode != 0 {
		return nil, errors.Wrapf(
			ErrRequestFailed,
			"[%s] execution failed",
			caller,
		)
	}

	return future, nil
}

type riskHub struct {
	UnimplementedRohonMonitorServer

	apiCache      sync.Map
	apiReqTimeout time.Duration
}

func (hub *riskHub) getApiInstance(idt string) (*riskApi, error) {
	if api, exist := hub.apiCache.Load(idt); exist {
		return api.(*riskApi), nil
	}

	return nil, errors.Wrapf(
		ErrRiskApiNotFound, "invalid risk api identity: %s",
		idt,
	)
}

func (hub *riskHub) Init(ctx context.Context, req *Request) (result *Result, err error) {
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

	hub.apiCache.Store(identity, &api)
	result.ReqId = -1
	result.Response = &Result_ApiIdentity{ApiIdentity: identity}

	return
}

func (hub *riskHub) ReqUserLogin(ctx context.Context, req *Request) (result *Result, err error) {
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var future rohon.Promise[rohon.RspUserLogin]
	if future, err = checkFuture(
		api.ins.AsyncReqUserLogin(&rohon.RiskUser{
			UserID:   req.GetLogin().UserId,
			Password: req.GetLogin().Password,
		}),
		"ReqUserLogin",
	); err != nil {
		return
	}

	if err = future.Then(func(r *rohon.Result[rohon.RspUserLogin]) error {
		login := <-r.Data

		var pri_type PrivilegeType

		switch login.PrivilegeType {
		case rohon.RH_MONITOR_ADMINISTRATOR:
			pri_type = admin
		case rohon.RH_MONITOR_NOMAL:
			pri_type = normal
		}

		result.Response = &Result_UserLogin{
			UserLogin: &RspUserLogin{
				UserId:        login.UserID,
				TradingDay:    login.TradingDay,
				LoginTime:     login.LoginTime,
				PrivilegeType: pri_type,
				PrivilegeInfo: login.InfoPrivilegeType,
			},
		}
		return nil
	}).Finally(
		reqFinalFn[rohon.RspUserLogin](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "wait rsp result failed")
	}

	return
}

func (hub *riskHub) ReqUserLogout(ctx context.Context, req *Request) (result *Result, err error) {
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var future rohon.Promise[rohon.RspUserLogout]
	if future, err = checkFuture(
		api.ins.AsyncReqUserLogout(),
		"ReqUserLogout",
	); err != nil {
		return
	}

	if err = future.Then(func(r *rohon.Result[rohon.RspUserLogout]) error {
		logout := <-r.Data

		result.Response = &Result_UserLogout{
			UserLogout: &RspUserLogout{
				UserId: logout.UserID,
			},
		}

		return nil
	}).Finally(
		reqFinalFn[rohon.RspUserLogout](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "wait rsp result failed")
	}

	return
}

func (hub *riskHub) ReqQryMonitorAccounts(ctx context.Context, req *Request) (result *Result, err error) {
	var api *riskApi
	if api, err = hub.getApiInstance(req.GetApiIdentity()); err != nil {
		return
	}

	var future rohon.Promise[rohon.Investor]
	if future, err = checkFuture(
		api.ins.AsyncReqQryMonitorAccounts(),
		"ReqQryMonitorAccounts",
	); err != nil {
		return
	}

	if err = future.Then(func(r *rohon.Result[rohon.Investor]) error {
		investors := InvestorList{}

		for inv := range r.Data {
			investors.Data = append(investors.Data, &Investor{
				BrokerId:   inv.BrokerID,
				InvestorId: inv.InvestorID,
			})
		}

		result.Response = &Result_Investors{
			Investors: &investors,
		}

		return nil
	}).Finally(
		reqFinalFn[rohon.Investor](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "wait rsp result failed")
	}

	return
}

func NewRohonMonitorHub(svr *grpc.Server) RohonMonitorServer {
	pb := &riskHub{}
	RegisterRohonMonitorServer(svr, pb)
	return pb
}

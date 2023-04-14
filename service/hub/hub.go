package hub

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/frozenpine/msgqueue/channel"
	"github.com/frozenpine/msgqueue/core"
	rohon "github.com/frozenpine/rhmonitor4go"
	rhapi "github.com/frozenpine/rhmonitor4go/api"
	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrInvalidArgs     = errors.New("invalid request args")
	ErrRiskApiNotFound = errors.New("risk api not found")
	ErrRequestFailed   = errors.New("request execution failed")
)

type RiskHub struct {
	service.UnimplementedRohonMonitorServer

	ctx             context.Context
	cancel          context.CancelFunc
	svr             *grpc.Server
	clientCache     sync.Map
	apiCache        sync.Map
	apiClientMapper sync.Map
	apiReqTimeout   time.Duration

	metrics promMetrics
}

func (hub *RiskHub) loadClient(idt string) (*client, error) {
	if c, exist := hub.clientCache.Load(idt); exist {
		return c.(*client), nil
	}

	return nil, errors.Wrapf(
		ErrClientNotFound,
		"invalid client source: %s", idt,
	)
}

func (hub *RiskHub) loadAndDelClient(idt string) (*client, error) {
	if c, exist := hub.clientCache.LoadAndDelete(idt); exist {
		return c.(*client), nil
	}

	return nil, errors.Wrapf(
		ErrClientNotFound,
		"invalid client source: %s", idt,
	)
}

func (hub *RiskHub) loadApiInstance(idt string) (*grpcRiskApi, error) {
	if api, exist := hub.apiCache.Load(idt); exist {
		return api.(*grpcRiskApi), nil
	}

	return nil, errors.Wrapf(
		ErrRiskApiNotFound,
		"invalid risk api identity: %s", idt,
	)
}

func (hub *RiskHub) loadAndDelApiInstance(idt string) (*grpcRiskApi, error) {
	if api, exist := hub.apiCache.LoadAndDelete(idt); exist {
		return api.(*grpcRiskApi), nil
	}

	return nil, errors.Wrapf(
		ErrRiskApiNotFound, "invalid risk api identity: %s",
		idt,
	)
}

func (hub *RiskHub) newClient(
	ctx context.Context,
	api *grpcRiskApi,
) *client {
	peer, _ := peer.FromContext(ctx)

	idt := uuid.NewV3(
		uuid.NamespaceDNS,
		fmt.Sprintf("%s://%s", peer.Addr.Network(), peer.Addr.String()),
	).String()

	c, _ := hub.clientCache.LoadOrStore(idt, &client{
		idt:  idt,
		peer: peer,
		api:  api,
	})

	hub.apiClientMapper.Store(frontToIdentity(api.front), ctx)

	return c.(*client)
}

func (hub *RiskHub) Init(ctx context.Context, req *service.Request) (result *service.Result, err error) {
	front := req.GetFront()
	if front == nil {
		err = errors.Wrap(ErrInvalidArgs, "[Init] func should set [RiskServer] arg")
		return
	}

	apiIdentity := frontToIdentity(front)

	var api *grpcRiskApi

	if api, err = hub.loadApiInstance(apiIdentity); err != nil {
		if errors.Is(err, ErrRiskApiNotFound) {
			api = &grpcRiskApi{
				broadcastCh:     make(chan string, broadcastBufferSize),
				broadcastBuffer: make([]string, 0, broadcastBufferSize),
				ctx:             context.Background(),
				front:           front,
			}

			if err = api.Init(
				front.BrokerId, front.ServerAddr,
				int(front.ServerPort), api,
			); err != nil {
				log.Printf("Create AsyncRHMonitorApi failed: %+v", err)

				return
			}

			hub.apiCache.Store(apiIdentity, api)
			err = nil

			log.Printf("New risk api created: %+v", front)
		} else {
			return
		}
	}

	c := hub.newClient(ctx, api)

	result = &service.Result{}

	result.ReqId = -1
	result.Response = &service.Result_ApiIdentity{ApiIdentity: c.idt}

	log.Printf("New client initiated: %s, %+v", c.idt, c)

	return
}

func (hub *RiskHub) Release(ctx context.Context, req *service.Request) (empty *emptypb.Empty, err error) {
	if _, err = hub.loadAndDelClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	empty = &emptypb.Empty{}

	log.Printf("Releasing risk api: %s", req.GetApiIdentity())

	return
}

func (hub *RiskHub) ReqUserLogin(ctx context.Context, req *service.Request) (result *service.Result, err error) {
	var c *client

	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	user := rohon.RiskUser{
		UserID:   req.GetLogin().UserId,
		Password: req.GetLogin().Password,
	}

	result = &service.Result{}

	// if err = c.checkConnect(); err != nil {
	// 	return
	// }

	if c.api.isLoggedIn() {
		result.ReqId = -1
		if user == c.api.state.user {
			c.login.Store(true)
			result.Response = &service.Result_UserLogin{
				UserLogin: service.ConvertRspLogin(&c.api.state.rspLogin),
			}

			log.Printf("Client login with api cache: %s", c)
		} else {
			err = errors.New("[grpc] Invalid username or password")
		}
	} else {
		var promise rhapi.Promise[rohon.RspUserLogin]

		if promise, err = checkPromise(
			c.api.AsyncReqUserLogin(&user),
			"ReqUserLogin",
		); err != nil {
			return
		}

		if err = promise.Then(func(r rhapi.Result[rohon.RspUserLogin]) error {
			login := <-r.GetData()

			result.Response = &service.Result_UserLogin{
				UserLogin: service.ConvertRspLogin(login),
			}

			c.login.Store(true)
			c.api.state.loggedIn.Store(true)
			c.api.state.user = user
			c.api.state.rspLogin = *login

			return nil
		}).Finally(
			reqFinalFn[rohon.RspUserLogin](result),
		).Await(ctx, hub.apiReqTimeout); err != nil {
			err = errors.Wrap(err, "[ReqUserLogin] wait rsp result failed")
		}
	}

	return
}

func (hub *RiskHub) ReqUserLogout(ctx context.Context, req *service.Request) (result *service.Result, err error) {
	var c *client
	if c, err = hub.loadAndDelClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	result = &service.Result{}

	if err = c.checkLogin(); err != nil {
		result.ReqId = -1
		log.Printf("Client not logged in: %+v", err)
		return
	}

	var promise rhapi.Promise[rohon.RspUserLogout]
	if promise, err = checkPromise(
		c.api.AsyncReqUserLogout(),
		"ReqUserLogout",
	); err != nil {
		return
	}

	if err = promise.Then(func(r rhapi.Result[rohon.RspUserLogout]) error {
		logout := <-r.GetData()

		result.Response = &service.Result_UserLogout{
			UserLogout: service.ConvertRspLogout(logout),
		}

		c.api.Release()
		hub.loadAndDelApiInstance(frontToIdentity(c.api.front))

		return nil
	}).Finally(
		reqFinalFn[rohon.RspUserLogout](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqUserLogout] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) ReqQryMonitorAccounts(ctx context.Context, req *service.Request) (result *service.Result, err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	result = &service.Result{}

	if err = c.checkLogin(); err != nil {
		result.ReqId = -1
		log.Printf("Client not logged in: %+v", err)
		return
	}

	var promise rhapi.Promise[rohon.Investor]
	if promise, err = checkPromise(
		c.api.AsyncReqQryMonitorAccounts(),
		"ReqQryMonitorAccounts",
	); err != nil {
		return
	}

	if err = promise.Then(func(r rhapi.Result[rohon.Investor]) error {
		investors := service.InvestorList{}

		for inv := range r.GetData() {
			investors.Data = append(investors.Data, service.ConvertInvestor(inv))
		}

		result.Response = &service.Result_Investors{
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

func (hub *RiskHub) ReqQryInvestorMoney(ctx context.Context, req *service.Request) (result *service.Result, err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	result = &service.Result{}

	if err = c.checkLogin(); err != nil {
		result.ReqId = -1
		log.Printf("Client not logged in: %+v", err)
		return
	}

	var promise rhapi.Promise[rohon.Account]

	if inv := req.GetInvestor(); inv != nil {
		investor := &rohon.Investor{
			BrokerID:   inv.BrokerId,
			InvestorID: inv.InvestorId,
		}

		if promise, err = checkPromise(
			c.api.AsyncReqQryInvestorMoney(investor),
			"ReqQryInvestorMoney",
		); err != nil {
			return
		}
	} else {
		if promise, err = checkPromise(
			c.api.AsyncReqQryAllInvestorMoney(),
			"ReqQryInvestorMoney",
		); err != nil {
			return
		}
	}

	if err = promise.Then(func(r rhapi.Result[rohon.Account]) error {
		accounts := &service.AccountList{}

		for acct := range r.GetData() {
			c.api.state.accountCache.LoadOrStore(acct.AccountID, acct)

			accounts.Data = append(accounts.Data, service.ConvertAccount(acct))
		}

		result.Response = &service.Result_Accounts{Accounts: accounts}

		return nil
	}).Finally(
		reqFinalFn[rohon.Account](result),
	).Await(ctx, hub.apiReqTimeout); err != nil {
		err = errors.Wrap(err, "[ReqQryInvestorMoney] wait rsp result failed")
	}

	return
}

func (hub *RiskHub) SubInvestorOrder(req *service.Request, stream service.RohonMonitor_SubInvestorOrderServer) (err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	if err = c.checkLogin(); err != nil {
		log.Printf("Client not logged in: %+v", err)
		return
	}

	p := channel.NewMemoChannel[*rohon.Order](
		c.api.ctx, frontToIdentity(c.api.front), 0,
	)

	if c.api.state.orderFlow.CompareAndSwap(nil, p) {
		result := c.api.AsyncReqSubAllInvestorOrder()

		if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
			err = errors.Wrap(err, "sub investor's order failed")

			return
		}

		go func() {
			for ord := range result.GetData() {
				p.Publish(ord, -1)
			}
		}()
	} else {
		p.Release()
		p = nil
	}

	ordFlow := c.api.state.orderFlow.Load()

	subID, ch := ordFlow.Subscribe(
		c.String(), core.Quick,
	)
	defer ordFlow.UnSubscribe(subID)

	filter := req.GetInvestor()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("Client[%s] disconnected.", c.idt)
			err = ordFlow.UnSubscribe(subID)
			return
		case ord := <-ch:
			if ord == nil {
				log.Print("Nil pointer in order flow")
				continue
			}

			if filter != nil && ord.AccountID != filter.InvestorId {
				continue
			}

			stream.Send(service.ConvertOrder(ord))
		}
	}
}

func (hub *RiskHub) SubInvestorTrade(req *service.Request, stream service.RohonMonitor_SubInvestorTradeServer) (err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	if err = c.checkLogin(); err != nil {
		log.Printf("Client not logged in: %+v", err)
		return
	}

	p := channel.NewMemoChannel[*rohon.Trade](
		c.api.ctx, frontToIdentity(c.api.front), 0,
	)

	if c.api.state.tradeFlow.CompareAndSwap(nil, p) {
		result := c.api.AsyncReqSubAllInvestorTrade()

		if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
			err = errors.Wrap(err, "sub investor's trade failed")

			return
		}

		go func() {
			for td := range result.GetData() {
				p.Publish(td, -1)
			}
		}()
	} else {
		p.Release()
		p = nil
	}

	tdFlow := c.api.state.tradeFlow.Load()

	subID, ch := tdFlow.Subscribe(
		c.String(), core.Quick,
	)
	defer tdFlow.UnSubscribe(subID)

	filter := req.GetInvestor()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("Client[%s] disconnected.", c.idt)
			err = tdFlow.UnSubscribe(subID)
			return
		case td := <-ch:
			if td == nil {
				log.Print("Nil pointer in trade flow")
				continue
			}

			if filter != nil && td.InvestorID != filter.InvestorId {
				continue
			}

			stream.Send(service.ConvertTrade(td))
		}
	}
}

func (hub *RiskHub) SubInvestorMoney(req *service.Request, stream service.RohonMonitor_SubInvestorMoneyServer) (err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	if err = c.checkLogin(); err != nil {
		log.Printf("Client not logged in: %+v", err)
		return
	}

	p := channel.NewMemoChannel[*rohon.Account](
		c.api.ctx, frontToIdentity(c.api.front), 0,
	)

	if c.api.state.accountFlow.CompareAndSwap(nil, p) {
		result := c.api.AsyncReqSubAllInvestorMoney()

		if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
			err = errors.Wrap(err, "sub investor's money failed")

			return
		}

		go func() {
			for acct := range result.GetData() {
				last, exist := c.api.state.accountCache.LoadOrStore(acct.AccountID, acct)

				if exist && service.RohonCompareAccount(last.(*rohon.Account), acct) {
					// 动态权益无变化, 不推送消息, 减少数据量
					log.Printf(
						"Account[%s] dynamic balance unchanged: %+v, %+v",
						acct.AccountID, last, acct,
					)
					continue
				}

				p.Publish(acct, -1)
			}
		}()
	} else {
		p.Release()
		p = nil
	}

	acctFlow := c.api.state.accountFlow.Load()

	subID, ch := acctFlow.Subscribe(
		c.String(), core.Quick,
	)
	defer acctFlow.UnSubscribe(subID)

	filter := req.GetInvestor()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("Client[%s] disconnected.", c.idt)
			err = acctFlow.UnSubscribe(subID)
			return
		case acct := <-ch:
			if acct == nil {
				log.Print("Nil pointer in account flow")
				continue
			}

			if filter != nil && filter.InvestorId != acct.AccountID {
				continue
			}

			stream.Send(service.ConvertAccount(acct))
		}
	}
}

func (hub *RiskHub) SubInvestorPosition(req *service.Request, stream service.RohonMonitor_SubInvestorPositionServer) (err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
		log.Printf("Client not found: %+v", err)
		return
	}

	if err = c.checkLogin(); err != nil {
		log.Printf("Client not logged in: %+v", err)
		return
	}

	p := channel.NewMemoChannel[*rohon.Position](
		c.api.ctx, frontToIdentity(c.api.front), 0,
	)

	if c.api.state.positionFlow.CompareAndSwap(nil, p) {
		result := c.api.AsyncReqSubAllInvestorPosition()

		if err = result.Await(stream.Context(), hub.apiReqTimeout); err != nil {
			err = errors.Wrap(err, "sub investor's postion failed")

			return
		}

		go func() {
			for pos := range result.GetData() {
				p.Publish(pos, -1)
			}
		}()
	} else {
		p.Release()
		p = nil
	}

	posFlow := c.api.state.positionFlow.Load()

	subID, ch := posFlow.Subscribe(
		c.String(), core.Quick,
	)
	defer posFlow.UnSubscribe(subID)

	filter := req.GetInvestor()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("Client[%s] disconnected.", c.idt)
			err = posFlow.UnSubscribe(subID)
			return
		case pos := <-ch:
			if pos == nil {
				log.Print("Nil pointer in position flow")
				continue
			}

			if filter != nil && filter.InvestorId != pos.InvestorID {
				continue
			}

			stream.Send(service.ConvertPosition(pos))
		}
	}
}

func (hub *RiskHub) SubBroadcast(req *service.Request, stream service.RohonMonitor_SubBroadcastServer) (err error) {
	var c *client
	if c, err = hub.loadClient(req.GetApiIdentity()); err != nil {
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

	for msg := range c.api.broadcastCh {
		stream.Send(&service.Broadcast{Message: msg})
	}

	return
}

func (hub *RiskHub) Serve(listen net.Listener) (err error) {
	go func() {
		if err = hub.svr.Serve(listen); err != nil {
			hub.cancel()
		}
	}()

	<-hub.ctx.Done()

	if err != nil {
		gracefulStop := make(chan struct{})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		log.Printf("Waiting 10 seconds to graceful stop grpc server.")
		go func() {
			hub.svr.GracefulStop()
			close(gracefulStop)
		}()

		select {
		case <-ctx.Done():
			log.Printf("Waiting timeout, stop grpc server forcelly.")
			hub.svr.Stop()
		case <-gracefulStop:
			log.Printf("Server stopped gracefully.")
		}
		cancel()
	}

	return
}

func NewRohonMonitorHub(ctx context.Context, tls *tls.Config) *RiskHub {
	var cancel context.CancelFunc

	if ctx == nil {
		ctx, cancel = signal.NotifyContext(
			context.Background(),
			os.Interrupt, os.Kill,
		)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	svr := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tls)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    5 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)

	pb := &RiskHub{
		ctx:    ctx,
		cancel: cancel,
		svr:    svr,
	}

	service.RegisterRohonMonitorServer(svr, pb)

	// reflection.Register(svr)

	return pb
}

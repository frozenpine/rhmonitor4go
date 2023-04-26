package hub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/frozenpine/msgqueue/channel"
	rohon "github.com/frozenpine/rhmonitor4go"
	rhapi "github.com/frozenpine/rhmonitor4go/api"
	"github.com/frozenpine/rhmonitor4go/service"
)

const broadcastBufferSize = 10

type apiState struct {
	connected atomic.Bool
	loggedIn  atomic.Bool
	user      rohon.RiskUser
	rspLogin  rohon.RspUserLogin

	orderFlow    atomic.Pointer[channel.MemoChannel[*rohon.Order]]
	tradeFlow    atomic.Pointer[channel.MemoChannel[*rohon.Trade]]
	positionFlow atomic.Pointer[channel.MemoChannel[*rohon.Position]]
	accountFlow  atomic.Pointer[channel.MemoChannel[*rohon.Account]]

	accountCache sync.Map
}

func frontToIdentity(f *service.RiskServer) string {
	if f == nil {
		return ""
	}

	return fmt.Sprintf("%s:%d@%s", f.ServerAddr, f.ServerPort, f.BrokerId)
}

type grpcRiskApi struct {
	rhapi.AsyncRHMonitorApi

	idt   string
	ctx   context.Context
	front *service.RiskServer
	// broadcastSub    atomic.Bool
	broadcastCh chan string
	// broadcastLock   sync.RWMutex
	broadcastBuffer []string

	state apiState

	clients sync.Map
}

func (api *grpcRiskApi) String() string {
	return fmt.Sprintf(
		"gRPCRhMonitorApi[%s] %s@%s:%d Connect[%v] Login[%v]",
		api.idt, api.state.user.UserID,
		api.front.ServerAddr, api.front.ServerPort,
		api.state.connected.Load(), api.state.loggedIn.Load(),
	)
}

func (api *grpcRiskApi) appendClient(c *client) *client {
	if c != nil {
		api.clients.Store(c.idt, c)
	}

	return c
}

func (api *grpcRiskApi) removeClient(c *client) *client {
	if c != nil {
		api.clients.Delete(c.idt)
	}

	return c
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

func (api *grpcRiskApi) isConnected() bool {
	return api.state.connected.Load()
}

func (api *grpcRiskApi) isLoggedIn() bool {
	return api.state.loggedIn.Load()
}

func (api *grpcRiskApi) OnFrontConnected() {
	api.HandleConnected()

	api.state.connected.Store(true)

	api.sendBroadcast(fmt.Sprintf(
		"Front[%s:%d] connected",
		api.front.ServerAddr, api.front.ServerPort,
	))
}

func (api *grpcRiskApi) OnFrontDisconnected(reason rohon.Reason) {
	api.HandleDisconnected()

	api.state.connected.Store(false)

	api.sendBroadcast(fmt.Sprintf(
		"Front[%s:%d] disconnected: %v",
		api.front.ServerAddr, api.front.ServerPort, reason,
	))
}

func (api *grpcRiskApi) OnRspUserLogin(
	login *rohon.RspUserLogin,
	info *rohon.RspInfo,
	reqID int64,
) {
	if info == nil || info.ErrorID == 0 {
		api.state.loggedIn.Store(true)
		api.state.rspLogin = *login
	}

	api.AsyncRHMonitorApi.OnRspUserLogin(login, info, reqID)
}

func reqFinalFn[T rhapi.RHRiskData](result *service.Result) rhapi.CallbackFn[T] {
	return func(req rhapi.Result[T]) error {
		result.ReqId = int32(req.GetRequestID())

		if rsp := req.GetRspInfo(); rsp != nil {
			result.RspInfo = &service.RspInfo{
				ErrorId:  int32(rsp.ErrorID),
				ErrorMsg: rsp.ErrorMsg,
			}
		}
		return nil
	}
}

func checkPromise[T rhapi.RHRiskData](result rhapi.Result[T], caller string) (rhapi.Promise[T], error) {
	return result, result.GetError()
}

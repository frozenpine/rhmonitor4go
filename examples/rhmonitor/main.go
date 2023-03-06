package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	rohon "github.com/frozenpine/rhmonitor4go"
)

var (
	// remoteAddr = "129.211.138.170"
	remoteAddr = "210.22.96.58"
	// remotePort = 20002
	remotePort = 11102
	// riskUser   = "rdcesfk"
	riskUser = "rdfk"
	riskPass = "888888"
	brokerID = "RohonDemo"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
}

// overwrite default api callback functions
// connected, disconnected, login, logout, investor events
// must call event handler to process flags
type myApi struct {
	rohon.RHMonitorApi
}

func (api *myApi) OnFrontConnected() {
	api.HandleConnected()
	log.Println("MyApi OnFrontConnected called.")
}

func (api *myApi) OnRspQryInvestorMoney(acct *rohon.Account, info *rohon.RspInfo, reqID int64, isLast bool) {
	log.Println(acct, info, reqID, isLast)
}

func main() {
	api := &myApi{}

	log.Printf("Instance: %p", api)

	if err := api.Init(brokerID, remoteAddr, remotePort, api); err != nil {
		log.Fatalf("Risk api init failed: +%v", err)
	}

	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	login := rohon.RiskUser{
		UserID:   riskUser,
		Password: riskPass,
	}

	var (
		reqID int64
		rtn   int
	)

	if reqID, rtn = api.ReqUserLogin(&login); rtn != 0 {
		log.Fatalf("ReqUserLogin[%d] faield: %d", reqID, rtn)
	}

	if reqID, rtn = api.ReqQryMonitorAccounts(); rtn != 0 {
		log.Fatalf("ReqQryMonitorAccounts[%d] failed: %d", reqID, rtn)
	}

	if rtn = api.ReqQryAllInvestorMoney(); rtn != 0 {
		log.Fatalf("ReqQryAllInvestorMoney failed: %d", rtn)
	}

	// if rtn = api.ReqQryAllInvestorPosition(); rtn != 0 {
	// 	log.Fatalf("ReqQryAllInvestorPosition failed: %d", rtn)
	// }

	// if rtn = api.ReqSubAllInvestorOrder(); rtn != 0 {
	// 	log.Fatalf("ReqSubAllInvestorOrder failed: %d", rtn)
	// }

	// if rtn = api.ReqSubAllInvestorTrade(); rtn != 0 {
	// 	log.Fatalf("ReqSubAllInvestorTrade failed: %d", rtn)
	// }

	<-ctx.Done()

	api.Release()
}

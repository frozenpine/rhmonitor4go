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

type myApi struct {
	rohon.RHMonitorApi
}

func (api *myApi) OnFrontConnected() {
	log.Println("MyApi OnFrontConnected called.")
	api.RHMonitorApi.OnFrontConnected()
}

func main() {
	api := &myApi{}
	// api := &rohon.RHMonitorApi{}

	log.Printf("Instance: %p", api)

	api.Init(brokerID, remoteAddr, remotePort, api)

	// if api == nil {
	// 	log.Fatal("Create api instance failed.")
	// }

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

	if rtn = api.ReqQryAllInvestorPosition(); rtn != 0 {
		log.Fatalf("ReqQryAllInvestorPosition failed: %d", rtn)
	}

	if rtn = api.ReqSubAllInvestorOrder(); rtn != 0 {
		log.Fatalf("ReqSubAllInvestorOrder failed: %d", rtn)
	}

	if rtn = api.ReqSubAllInvestorTrade(); rtn != 0 {
		log.Fatalf("ReqSubAllInvestorTrade failed: %d", rtn)
	}

	<-ctx.Done()
}

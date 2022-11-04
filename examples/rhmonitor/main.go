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

func main() {
	api := rohon.NewRHMonitorApi(brokerID, remoteAddr, remotePort)

	if api == nil {
		log.Fatal("Create api instance failed.")
	}

	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	login := rohon.RiskUser{
		UserID:   riskUser,
		Password: riskPass,
	}

	api.ReqUserLogin(&login)

	api.ReqQryMonitorAccounts()

	api.ReqQryAllInvestorMoney()

	api.ReqQryAllInvestorPosition()

	// api.ReqSubAllInvestorOrder()

	api.ReqSubAllInvestorTrade()

	<-ctx.Done()
}

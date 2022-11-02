package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/frozenpine/cRiskApi/rohon"
)

var (
	remoteAddr = "129.211.138.170"
	remotePort = 20002
	riskUser   = "rdcesfk"
	riskPass   = "888888"
	brokerID   = "RohonDemo"
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

	<-ctx.Done()
}

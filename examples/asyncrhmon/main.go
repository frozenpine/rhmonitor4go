package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

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
	api := rohon.NewAsyncRHMonitorApi(brokerID, remoteAddr, remotePort)

	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	login := rohon.RiskUser{
		UserID:   riskUser,
		Password: riskPass,
	}

	var err error
	var timeout = time.Second * 10

	if err = api.AsyncReqUserLogin(&login).Then(
		func(r rohon.Result[rohon.RspUserLogin]) error {
			login := <-r.GetData()

			log.Printf("Risk user logged in: %+v", login)

			// return errors.New("test error")
			return nil
		},
	).Catch(
		func(r rohon.Result[rohon.RspUserLogin]) error {
			log.Print("Catch called.")
			return nil
		},
	).Finally(
		func(r rohon.Result[rohon.RspUserLogin]) error {
			log.Print("Final called.")
			return nil
		},
	).Await(ctx, timeout); err != nil {
		log.Fatalf("AsyncReqUserLogin failed: %+v", err)
	}

	if err = api.AsyncReqQryMonitorAccounts().Then(
		func(r rohon.Result[rohon.Investor]) error {
			for inv := range r.GetData() {
				fmt.Printf("AsyncOnRspQryMonitorAccounts: %+v", inv)
			}
			return nil
		},
	).Await(ctx, timeout); err != nil {
		log.Fatalf("AyncReqQryMonitorAccounts failed: %+v", err)
	}

	<-ctx.Done()
}

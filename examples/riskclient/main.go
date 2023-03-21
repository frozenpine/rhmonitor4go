package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/frozenpine/rhmonitor4go/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rpcAddr     = ""
	rpcPort     = 1234
	clientCert  = "rhmonitor.crt"
	clientKey   = "rhmonitor.key"
	ca          = "ca.crt"
	timeout     = 5
	riskSvr     = ""
	riskSvrConn = regexp.MustCompile("tcp://([0-9.]+):([0-9]+)")
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	flag.StringVar(&rpcAddr, "addr", rpcAddr, "gRPC remote addr")
	flag.IntVar(&rpcPort, "port", rpcPort, "gRPC remote port")
	flag.StringVar(&clientCert, "cert", clientCert, "gRPC client cert path")
	flag.StringVar(&clientKey, "key", clientKey, "gRPC client cert key path")
	flag.StringVar(&ca, "ca", ca, "gRPC server cert CA path")
	flag.IntVar(&timeout, "timeout", timeout, "gRPC call deadline in second")
	flag.StringVar(&riskSvr, "svr", riskSvr, "Rohon risk server conn in format: tcp://{addr}:{port}")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	match := riskSvrConn.FindStringSubmatch(riskSvr)
	if len(match) != 2 {
		log.Fatalf("Invalid risk server conn: %s", riskSvr)
	}
	riskSvrAddr := match[0]
	riskSvrPort, _ := strconv.Atoi(match[1])

	log.Printf("Loading gRPC client cert pair")
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("Load RPC cert pair failed: %+v", err)
	}

	caPool := x509.NewCertPool()
	log.Printf("Loading gRPC client CA cert")
	caData, err := os.ReadFile(ca)
	if err != nil {
		log.Fatalf("Load gRPC cert CA failed: %+v", err)
	}
	if ok := caPool.AppendCertsFromPEM(caData); !ok {
		log.Fatalf("Parse gRPC cert CA failed: %s", ca)
	}

	tlsConfig := &tls.Config{
		ServerName:   rpcAddr,
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	remoteAddr := fmt.Sprintf("%s:%d", rpcAddr, rpcPort)
	log.Printf("Connecting to gRPC server: %s", remoteAddr)
	conn, err := grpc.DialContext(
		ctx, remoteAddr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(tlsConfig),
		),
	)
	// conn, err := grpc.DialContext(
	// 	ctx, remoteAddr,
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// )
	if err != nil {
		log.Fatalf("Connet to gRPC server[%s] failed: %+v", remoteAddr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Fail to close gRPC conn: %+v", err)
		} else {
			log.Println("gRPC conn closed")
		}
	}()

	// cli := CLI{}
	client := service.NewRohonMonitorClient(conn)

	var (
		result      *service.Result
		apiIdentity string
		deadline    context.Context
		cancel      context.CancelFunc
	)

	deadline, cancel = context.WithTimeout(ctx, time.Second*time.Duration(timeout))

	if result, err = client.Init(deadline, &service.Request{
		Request: &service.Request_Front{
			Front: &service.RiskServer{
				ServerAddr: riskSvrAddr,
				ServerPort: int32(riskSvrPort),
			},
		},
	}); err != nil {
		log.Fatalf("Init remote risk api failed: %+v", err)
	} else {
		apiIdentity = result.GetApiIdentity()

		log.Printf("Remote risk api initiated: %s", apiIdentity)
	}
	cancel()
	defer func() {
		if _, err = client.Release(ctx, &service.Request{
			ApiIdentity: apiIdentity,
		}); err != nil {
			log.Fatalf("Release api failed: %+v", err)
		}
	}()

	var broadcast service.RohonMonitor_SubBroadcastClient
	if broadcast, err = client.SubBroadcast(ctx, &service.Request{
		ApiIdentity: apiIdentity,
	}); err != nil {
		log.Printf("Sub broadcast failed: %+v", err)
	} else {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if msg, err := broadcast.Recv(); err != nil {
						log.Printf("Receive broadcast failed: %+v", err)
					} else {
						log.Printf("Broadcast received: %s", msg.Message)
					}
				}
			}
		}()
	}

	deadline, cancel = context.WithTimeout(ctx, time.Second*time.Duration(timeout))

	if result, err = client.ReqUserLogin(deadline, &service.Request{
		ApiIdentity: apiIdentity,
		Request: &service.Request_Login{
			Login: &service.RiskUser{
				UserId:   "rdfk",
				Password: "888888",
			},
		},
	}); err != nil {
		log.Fatalf("Remote login failed: %+v", err)
	} else {
		log.Printf("Remote login: %+v", result.GetUserLogin())
	}
	cancel()

	deadline, cancel = context.WithTimeout(ctx, time.Second*time.Duration(timeout))

	if result, err = client.ReqQryMonitorAccounts(deadline, &service.Request{
		ApiIdentity: apiIdentity,
		// Request: &service.Request_Investor{
		// 	Investor: &service.Investor{InvestorId: ""},
		// },
	}); err != nil {
		log.Fatalf("Query accounts failed: +%v", err)
	} else {
		investors := result.GetInvestors()

		for _, inv := range investors.Data {
			fmt.Printf("gRPC query investor: %+v\n", inv)
		}
	}
	cancel()

	stream, err := client.SubInvestorMoney(ctx, &service.Request{
		ApiIdentity: apiIdentity,
	})
	if err != nil {
		log.Fatalf("Subscribe investor's account failed: %+v", err)
	}

	for {
		acct, err := stream.Recv()

		if err != nil {
			log.Printf("Receive investor's account failed: %+v", err)
			break
		}

		fmt.Printf("OnRtnInvestorMoney %+v", acct)
	}

	// log.Printf("Starting gRPC client")
	// if err := cli.Serve(ctx, service.NewRohonMonitorClient(conn)); err != nil {
	// 	log.Fatalf("Client running failed: %+v", err)
	// }
}

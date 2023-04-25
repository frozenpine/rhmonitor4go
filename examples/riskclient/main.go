package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/frozenpine/rhmonitor4go/service/client"
	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/proto"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type AccountList []string

func (l *AccountList) Set(value string) error {
	if value == "" {
		l = nil
		return nil
	}

	for _, v := range strings.Split(value, ",") {
		d := strings.TrimSpace(v)

		if d == "" {
			continue
		}

		*l = append(([]string)(*l), d)
	}

	return nil
}

func (l AccountList) String() string {
	return strings.Join(([]string)(l), ", ")
}

var (
	rpcAddr    = ""
	rpcPort    = 1234
	clientCert = "riskclient.crt"
	clientKey  = "riskclient.key"
	ca         = "ca.crt"
	timeout    = 5

	dbConn = "postgres://host=localhost port=5432 user=trade password=trade dbname=lingma"

	riskSvr        = ""
	riskSvrPattern = regexp.MustCompile("tcp://([a-zA-Z0-9]+)#(.+)@([0-9.]+):([0-9]+)")

	redisSvr        = "localhost:6379#1"
	redisSvrPattern = regexp.MustCompile("(?:(.+)@)?([a-z0-9A-Z.:].+)#([0-9]+)")
	redisChanBase   = "rohon.risk.accounts"
	redisFormat     = client.MsgPack

	barDuration = time.Minute

	accounts AccountList

	version, goVersion, gitVersion, buildTime string
	showVersion                               = false
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile)

	flag.StringVar(&rpcAddr, "addr", rpcAddr, "gRPC remote addr")
	flag.IntVar(&rpcPort, "port", rpcPort, "gRPC remote port")
	flag.StringVar(&clientCert, "cert", clientCert, "gRPC client cert path")
	flag.StringVar(&clientKey, "key", clientKey, "gRPC client cert key path")
	flag.StringVar(&ca, "ca", ca, "gRPC server cert CA path")
	flag.IntVar(&timeout, "timeout", timeout, "gRPC call deadline in second")

	flag.StringVar(&riskSvr, "svr", riskSvr, "Rohon risk server conn in format: tcp://{user}#{pass}@{addr}:{port}")

	flag.StringVar(&dbConn, "db", dbConn, "Database conn string")

	flag.StringVar(&redisSvr, "redis", redisSvr, "Redis server conn in format: ({pass}@)?{addr}:{port}#{db}")
	flag.StringVar(&redisChanBase, "chan", redisChanBase, "Redis publish base channel")
	flag.Var(&redisFormat, "format", "Redis message marshal format (default: proto3)")

	flag.Var(&accounts, "accounts", "Account list for filting")

	flag.DurationVar(&barDuration, "bar", barDuration, "Bar duration")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if showVersion {
		fmt.Printf(
			"Version: %s, Commit: %s, Build: %s @ %s\n",
			version, gitVersion, buildTime, goVersion,
		)
		os.Exit(0)
	}

	db, err := client.InitDB(dbConn)
	if err != nil {
		log.Fatal("Init db failed:", err)
	}
	defer db.Close()

	match := riskSvrPattern.FindStringSubmatch(riskSvr)
	if len(match) != 5 {
		log.Fatalf("Invalid risk server conn: %s", riskSvr)
	}
	riskUser := match[1]
	riskPass := match[2]
	riskSvrAddr := match[3]
	riskSvrPort, _ := strconv.Atoi(match[4])

	match = redisSvrPattern.FindStringSubmatch(redisSvr)
	if len(match) != 4 {
		log.Fatalf("Invalid redis server conn: %s", redisSvr)
	}
	redisSvrAddr := match[2]
	redisSvrPass := match[1]
	redisDB, _ := strconv.Atoi(match[3])

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

	rootCtx, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	remoteAddr := fmt.Sprintf("%s:%d", rpcAddr, rpcPort)
	log.Printf("Connecting to gRPC server: %s", remoteAddr)
	conn, err := grpc.DialContext(
		rootCtx, remoteAddr,
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisSvrAddr,
		Password: redisSvrPass,
		DB:       redisDB,
	})

	remote := service.NewRohonMonitorClient(conn)

	var (
		result      *service.Result
		apiIdentity string
		deadline    context.Context
		cancel      context.CancelFunc
	)

	deadline, cancel = context.WithTimeout(rootCtx, time.Second*time.Duration(timeout))

	if result, err = remote.Init(deadline, &service.Request{
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
		if _, err = remote.Release(rootCtx, &service.Request{
			ApiIdentity: apiIdentity,
		}); err != nil {
			log.Fatalf("Release api failed: %+v", err)
		}
	}()

	var broadcast service.RohonMonitor_SubBroadcastClient
	if broadcast, err = remote.SubBroadcast(rootCtx, &service.Request{
		ApiIdentity: apiIdentity,
	}); err != nil {
		log.Printf("Sub broadcast failed: %+v", err)
	} else {
		go func() {
			for {
				select {
				case <-rootCtx.Done():
					return
				default:
					if msg, err := broadcast.Recv(); err != nil {
						log.Printf("Receive broadcast failed: %+v", err)
						rootCancel()
						return
					} else {
						log.Printf("Broadcast received: %s", msg.Message)
					}
				}
			}
		}()
	}

	deadline, cancel = context.WithTimeout(rootCtx, time.Second*time.Duration(timeout))

	if result, err = remote.ReqUserLogin(deadline, &service.Request{
		ApiIdentity: apiIdentity,
		Request: &service.Request_Login{
			Login: &service.RiskUser{
				UserId:   riskUser,
				Password: riskPass,
			},
		},
	}); err != nil {
		log.Fatalf("Remote login failed: %+v", err)
	} else {
		log.Printf("Remote login: %+v", result.GetUserLogin())
	}
	cancel()

	deadline, cancel = context.WithTimeout(rootCtx, time.Second*time.Duration(timeout))

	if result, err = remote.ReqQryMonitorAccounts(deadline, &service.Request{
		ApiIdentity: apiIdentity,
	}); err != nil {
		log.Fatalf("Query accounts failed: +%v", err)
	} else {
		for _, inv := range result.GetInvestors().GetData() {
			fmt.Printf("gRPC query investor: %+v\n", inv)
		}
	}
	cancel()

	deadline, cancel = context.WithTimeout(rootCtx, time.Second*time.Duration(timeout))
	settAccounts := make(map[string]*service.Account)

	if result, err = remote.ReqQryInvestorMoney(deadline, &service.Request{
		ApiIdentity: apiIdentity,
		// Request: &service.Request_Investor{
		// 	Investor: &service.Investor{InvestorId: "lmhx01"},
		// },
	}); err != nil {
		log.Fatalf("Query investor's money failed: +%v", err)
	} else {
		for _, acct := range result.GetAccounts().GetData() {
			settAccounts[acct.Investor.InvestorId] = acct
			fmt.Printf("Queried account: %+v\n", acct)
		}
	}
	cancel()

	stream, err := remote.SubInvestorMoney(rootCtx, &service.Request{
		ApiIdentity: apiIdentity,
		// Request: &service.Request_Investor{
		// 	Investor: &service.Investor{InvestorId: "lmhx01"},
		// },
	})
	if err != nil {
		log.Fatalf("Subscribe investor's account failed: %+v", err)
	}

	var (
		buffer     []byte
		pubChan    = []string{redisChanBase, ""}
		marshaller func(any) ([]byte, error)
	)

	sinker, err := client.NewAccountSinker(rootCtx, client.Continuous, barDuration, settAccounts, stream)
	if err != nil {
		log.Fatalf("Create sinker failed: %+v", err)
	}

	switch redisFormat {
	case client.MsgProto3:
		marshaller = func(value any) ([]byte, error) {
			data, ok := value.(proto.Message)
			if !ok {
				return nil, errors.New("invalid proto3 message")
			}

			return proto.Marshal(data)
		}
	case client.MsgJson:
		marshaller = json.Marshal
	case client.MsgPack:
		marshaller = msgpack.Marshal
	default:
		log.Fatal("Unsupported message format: ", redisFormat)
	}

	for {
		select {
		case <-rootCtx.Done():
			return
		case acct := <-sinker.Data():
			if buffer, err = marshaller(acct); err != nil {
				log.Printf("Marshal account message failed: %s", err)
				continue
			}

			pubChan[1] = acct.InvestorID
			cmd := rdb.Publish(rootCtx, strings.Join(pubChan, "."), buffer)

			if err = cmd.Err(); err != nil {
				log.Printf(
					"Publish to redis[%s@%d] faield: %s",
					redisSvr, redisDB, err,
				)
			}
		}
	}
}

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
	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rpcAddr    = ""
	rpcPort    = 1234
	clientCert = "riskclient.crt"
	clientKey  = "riskclient.key"
	ca         = "ca.crt"
	timeout    = 5

	dbFile = "trade.db"

	riskSvr        = ""
	riskSvrPattern = regexp.MustCompile("tcp://([a-zA-Z0-9]+)#(.+)@([0-9.]+):([0-9]+)")

	redisSvr        = "localhost:6379#1"
	redisSvrPattern = regexp.MustCompile("(?:(.+)@)?([a-z0-9A-Z.:].+)#([0-9]+)")
	redisChanBase   = "rohon.risk.accounts"
	redisFormat     = MsgProto3

	barDuration = time.Minute
)

type MsgFormat uint8

func (msgFmt *MsgFormat) Set(value string) error {
	switch value {
	case "proto3":
		*msgFmt = MsgProto3
	case "json":
		*msgFmt = MsgJson
	case "msgpack":
		*msgFmt = MsgPack
	default:
		return errors.New("invalid msg format")
	}

	return nil
}

//go:generate stringer -type MsgFormat -linecomment
const (
	MsgProto3 MsgFormat = iota // proto3
	MsgJson                    // json
	MsgPack                    // msgpack
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	flag.StringVar(&rpcAddr, "addr", rpcAddr, "gRPC remote addr")
	flag.IntVar(&rpcPort, "port", rpcPort, "gRPC remote port")
	flag.StringVar(&clientCert, "cert", clientCert, "gRPC client cert path")
	flag.StringVar(&clientKey, "key", clientKey, "gRPC client cert key path")
	flag.StringVar(&ca, "ca", ca, "gRPC server cert CA path")
	flag.IntVar(&timeout, "timeout", timeout, "gRPC call deadline in second")

	flag.StringVar(&riskSvr, "svr", riskSvr, "Rohon risk server conn in format: tcp://{user}#{pass}@{addr}:{port}")

	flag.StringVar(&dbFile, "db", dbFile, "Database file")

	flag.StringVar(&redisSvr, "redis", redisSvr, "Redis server conn in format: ({pass}@)?{addr}:{port}#{db}")
	flag.StringVar(&redisChanBase, "chan", redisChanBase, "Redis publish base channel")
	flag.Var(&redisFormat, "format", "Redis message marshal format (default: proto3)")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if err := initDB(); err != nil {
		log.Fatal("Init db failed:", err)
	}

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

	var runCancel context.CancelFunc
	ctx := context.Background()
	ctx, runCancel = signal.NotifyContext(ctx, os.Interrupt, os.Kill)

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

	sinkAccount, err := NewSinkAccount(ctx)
	if err != nil {
		log.Fatal("Make sink handler failed:", err)
	}
	defer func() {
		db.Close()
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisSvrAddr,
		Password: redisSvrPass,
		DB:       redisDB,
	})

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

	var (
		buffer         []byte
		pubChan        = []string{redisChanBase, ""}
		defaultPubChan = redisChanBase + ".default"
		marshaller     func(any) ([]byte, error)
		receiveCh      = make(chan *service.Account, 1)
	)

	go func() {
		defer runCancel()

		for {
			acct, err := stream.Recv()

			if err != nil {
				log.Printf("Receive investor's account failed: %+v", err)
				break
			}

			fmt.Printf("OnRtnInvestorMoney %+v\n", acct)
			receiveCh <- acct
		}
	}()

	switch redisFormat {
	case MsgProto3:
		marshaller = func(value any) ([]byte, error) {
			data, ok := value.(proto.Message)
			if !ok {
				return nil, errors.New("invalid proto3 message")
			}

			return proto.Marshal(data)
		}
	case MsgJson:
		marshaller = json.Marshal
	default:
		log.Fatal("Unsupported message format: ", redisFormat)
	}

	now := time.Now()
	nextTs := now.Round(barDuration)

	if nextTs.Before(now) {
		nextTs = nextTs.Add(barDuration)
	}

	timer := time.NewTimer(nextTs.Sub(now))
	ticker := time.NewTicker(barDuration)
	ticker.Stop()
	waterMark := &service.Account{}
	var cmd *redis.IntCmd

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-timer.C:
			log.Printf("Bar ticker first initialized: %+v", t)
			ticker.Reset(barDuration)
			timer.Stop()

			waterMark.Timestamp = t.UnixMicro()
			buffer, err = waterMark.Marshal()
			if err != nil {
				log.Fatal("Can not marshal water mark")
			}
			cmd = rdb.Publish(ctx, defaultPubChan, buffer)
		case t := <-ticker.C:
			waterMark.Timestamp = t.UnixMicro()
			buffer, err = waterMark.Marshal()
			if err != nil {
				log.Fatal("Can not marshal water mark")
			}
			cmd = rdb.Publish(ctx, defaultPubChan, buffer)
		case acct := <-receiveCh:
			sinkAccount.FromAccount(acct)

			if err = sinkAccount.Insert(); err != nil {
				if errors.Is(err, ErrSameData) {
					continue
				} else {
					log.Fatal("Sink accout to db failed:", err)
				}
			}

			if buffer, err = marshaller(acct); err != nil {
				log.Printf("Marshal account message failed: %s", err)
				continue
			}

			pubChan[1] = acct.Investor.InvestorId
			cmd = rdb.Publish(ctx, strings.Join(pubChan, "."), buffer)
		}

		if err = cmd.Err(); err != nil {
			log.Printf(
				"Publish to redis[%s@%d] faield: %s",
				redisSvr, redisDB, err,
			)
		}
	}
}

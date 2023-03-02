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

	"github.com/frozenpine/rhmonitor4go/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rpcAddr    = ""
	rpcPort    = 1234
	clientCert = "rhmonitor.crt"
	clientKey  = "rhmonitor.key"
	ca         = "ca.crt"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	flag.StringVar(&rpcAddr, "addr", rpcAddr, "gRPC remote addr")
	flag.IntVar(&rpcPort, "port", rpcPort, "gRPC remote port")
	flag.StringVar(&clientCert, "cert", clientCert, "gRPC client cert path")
	flag.StringVar(&clientKey, "key", clientKey, "gRPC client cert key path")
	flag.StringVar(&ca, "ca", ca, "gRPC server cert CA path")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("Load RPC cert pair failed: %+v", err)
	}

	caPool := x509.NewCertPool()
	caData, err := os.ReadFile(ca)
	if err != nil {
		log.Fatalf("Load gRPC cert CA failed: %+v", err)
	}
	if ok := caPool.AppendCertsFromPEM(caData); !ok {
		log.Fatalf("Parse gRPC cert CA failed: %s", ca)
	}

	tlsConfig := &tls.Config{
		ServerName:   "riskhub.lingma.vip",
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	remoteAddr := fmt.Sprintf("%s:%d", rpcAddr, rpcPort)
	conn, err := grpc.DialContext(
		ctx, remoteAddr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(tlsConfig),
		),
	)
	if err != nil {
		log.Fatalf("Connet to gRPC server[%s] failed: %+v", remoteAddr, err)
	}
	defer conn.Close()

	cli := CLI{}

	if err := cli.Serve(ctx, service.NewRohonMonitorClient(conn)); err != nil {
		log.Fatalf("Client running failed: %+v", err)
	}
}

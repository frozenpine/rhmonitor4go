package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/frozenpine/rhmonitor4go/service/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	rpcAddr = "0.0.0.0"
	rpcPort = 1234
	svrCert = "riskhub.crt"
	svrKey  = "riskhub.key"
	ca      = "ca.crt"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	flag.StringVar(&rpcAddr, "addr", rpcAddr, "gRPC listen addr")
	flag.IntVar(&rpcPort, "port", rpcPort, "gRPC listen port")
	flag.StringVar(&svrCert, "cert", svrCert, "gRPC server cert path")
	flag.StringVar(&svrKey, "key", svrKey, "gRPC server cert key path")
	flag.StringVar(&ca, "ca", ca, "gRPC server cert CA path")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	listenAddr := fmt.Sprintf("%s:%d", rpcAddr, rpcPort)
	log.Printf("Binding gRPC listener: %s", listenAddr)
	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		log.Fatalf("Bind gRPC listening[tcp://%s] failed: %+v", listenAddr, err)
	}

	log.Printf("Loading gRPC server cert pair")
	cert, err := tls.LoadX509KeyPair(svrCert, svrKey)
	if err != nil {
		log.Fatalf("Load RPC cert pair failed: %+v", err)
	}

	caPool := x509.NewCertPool()
	log.Printf("Loading gRPC server CA cert")
	caData, err := os.ReadFile(ca)
	if err != nil {
		log.Fatalf("Load gRPC cert CA failed: %+v", err)
	}
	if ok := caPool.AppendCertsFromPEM(caData); !ok {
		log.Fatalf("Parse gRPC cert CA failed: %s", ca)
	}

	tlsConfig := tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
	}

	grpcSvr := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(&tlsConfig)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    5 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	// grpcSvr := grpc.NewServer()
	hub.NewRohonMonitorHub(grpcSvr)

	log.Printf("Starting gRPC server")
	if err := grpcSvr.Serve(listen); err != nil {
		log.Fatalf("Starting gRPC failed: %+v", err)
	}
}

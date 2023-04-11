module github.com/frozenpine/rhmonitor4go

go 1.19

require github.com/pkg/errors v0.9.1

require golang.org/x/text v0.7.0

require (
	github.com/frozenpine/msgqueue v0.0.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gofrs/uuid v4.3.1+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/mattn/go-sqlite3 v1.14.16
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
)

module github.com/lowRISC/opentitan-provisioning

go 1.18

replace github.com/lowRISC/opentitan-provisioning => ./

require (
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/miekg/pkcs11 v1.0.3
	go.etcd.io/etcd v3.3.27+incompatible
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 v3.5.1
	google.golang.org/grpc v1.41.0
	github.com/google/tink/go v1.6.1
	golang.org/x/crypto v0.0.0-20220307211146-efcb8507fb70
	golang.org/x/sync v0.1.0
	golang.org/x/tools v0.1.10
	github.com/cenkalti/backoff/v4 v4.2.1 
	github.com/elastic/elastic-transport-go/v8 v8.3.0 
	github.com/elastic/go-elasticsearch/v8 v8.10.1
)

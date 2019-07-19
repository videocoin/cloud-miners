module github.com/videocoin/cloud-miners

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/JensRantil/graphite-client v0.0.0-20151206234601-d93bf4b72f5a
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190715232110-2b613d287457 // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/grpc-gateway v1.9.4 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/videocoin/cloud-api v0.1.160
	github.com/videocoin/cloud-pkg v0.0.2
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610 // indirect
	google.golang.org/grpc v1.22.0
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

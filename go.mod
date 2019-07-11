module github.com/videocoin/cloud-miners

go 1.12

require (
	github.com/JensRantil/graphite-client v0.0.0-20151206234601-d93bf4b72f5a
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/jinzhu/gorm v1.9.10
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/videocoin/cloud-api v0.1.136
	github.com/videocoin/cloud-pkg v0.0.2
	go.uber.org/atomic v1.4.0 // indirect
	google.golang.org/grpc v1.22.0
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

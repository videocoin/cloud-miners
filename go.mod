module github.com/videocoin/cloud-miners

go 1.12

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/JensRantil/graphite-client v0.0.0-20151206234601-d93bf4b72f5a // indirect
	github.com/Pallinder/sillyname-go v0.0.0-20130730142914-97aeae9e6ba1 // indirect
	github.com/bradfitz/slice v0.0.0-20180809154707-2b758aa73013 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/google/uuid v1.1.1
	github.com/goombaio/namegenerator v0.0.0-20181006234301-989e774b106e
	github.com/grpc-ecosystem/grpc-gateway v1.11.3
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/lib/pq v1.2.0 // indirect
	github.com/mailru/dbr v3.0.0+incompatible
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/reality-lab-networks/liveplanet-api v0.0.0-20190906141833-b7fe3c9f4f36 // indirect
	github.com/segmentio/ksuid v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/videocoin/cloud-api v0.2.14
	github.com/videocoin/cloud-dispatcher v0.1.3 // indirect
	github.com/videocoin/cloud-pkg v0.0.6
	go4.org v0.0.0-20200312051459-7028f7b4a332 // indirect
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	google.golang.org/grpc v1.27.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

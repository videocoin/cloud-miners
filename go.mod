module github.com/videocoin/cloud-miners

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/AlekSi/pointer v1.1.0
	github.com/JensRantil/graphite-client v0.0.0-20151206234601-d93bf4b72f5a // indirect
	github.com/Pallinder/sillyname-go v0.0.0-20130730142914-97aeae9e6ba1 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/goombaio/namegenerator v0.0.0-20181006234301-989e774b106e
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
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	google.golang.org/grpc v1.23.0
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

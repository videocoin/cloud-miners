module github.com/videocoin/cloud-miners

go 1.14

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/uuid v1.1.1
	github.com/goombaio/namegenerator v0.0.0-20181006234301-989e774b106e
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/mailru/dbr v3.0.0+incompatible
	github.com/opentracing/opentracing-go v1.1.0
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.6.0
	github.com/videocoin/cloud-api v0.0.17
	github.com/videocoin/cloud-pkg v0.0.5
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f
	google.golang.org/grpc v1.29.1
)

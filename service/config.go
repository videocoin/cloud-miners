package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string        `envconfig:"-"`
	Version string        `envconfig:"-"`
	Logger  *logrus.Entry `envconfig:"-"`

	Addr            string `default:"0.0.0.0:5011"`
	MetricsAddr     string `default:"0.0.0.0:15011"`
	EmitterRPCAddr  string `envconfig:"EMITTER_RPC_ADDR" default:"0.0.0.0:5003"`
	IamEndpoint     string `envconfig:"IAM_ENDPOINT" default:"https://iam.dev.videocoinapis.com"`
	DBURI           string `default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI           string `envconfig:"MQURI" default:"amqp://guest:guest@127.0.0.1:5672"`
	AuthTokenSecret string `envconfig:"AUTH_TOKEN_SECRET" default:"secret"`
}

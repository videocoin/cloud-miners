package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	Addr        string `default:"0.0.0.0:5011"`
	MetricsAddr string `default:"0.0.0.0:15011"`
	DBURI       string `default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI       string `default:"amqp://guest:guest@127.0.0.1:5672" envconfig:"MQURI"`

	AuthTokenSecret string `default:"secret" envconfig:"AUTH_TOKEN_SECRET"`

	Logger *logrus.Entry `envconfig:"-"`
}

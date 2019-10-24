package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	Addr      string `default:"0.0.0.0:5011"`
	DBURI     string `default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	CarbonURL string `default:"http://localhost:8080/render" envconfig:"CARBONAPI_URL"`

	AuthTokenSecret string `default:"secret" envconfig:"AUTH_TOKEN_SECRET"`

	Logger *logrus.Entry `envconfig:"-"`
}

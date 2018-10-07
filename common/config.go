package common

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ClientId           string
	ClientSecret       string
	ClientRedirectPort string
	ClientServiceURL   string

	ServerPort         string
}

func GetConfig() (*Config, error) {
	c := &Config{}
	return c, envconfig.Process("awsspotboxes", c)
}
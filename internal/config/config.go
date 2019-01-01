package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

var (
	cfg Config
)

func init() {
	err := envconfig.Process("quimby", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Config holds env vars
type Config struct {
	InternalAddress string `required:"true"`
	BlockKey        string `required:"true"`
	HashKey         string `required:"true"`
	TLSCert         string `required:"true"`
	TLSKey          string `required:"true"`
}

func Get() Config {
	return cfg
}

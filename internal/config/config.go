package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

var (
	cfg *Config
)

// Config holds env vars
type Config struct {
	InternalAddress string `required:"true"`
	Host            string `required:"true"`
	BlockKey        string `required:"true"`
	HashKey         string `required:"true"`
	TLSCert         string `required:"true"`
	TLSKey          string `required:"true"`
}

func Get() Config {
	if cfg == nil {
		var c Config
		err := envconfig.Process("quimby", &c)
		if err != nil {
			log.Fatal(err.Error())
		}
		cfg = &c
	}
	return *cfg
}

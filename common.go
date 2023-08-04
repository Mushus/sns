package activitypub

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SoftwareName string `envconfig:"SOFTWARE_NAME" default:"activitypub"`
	Host         string `envconfig:"HOST" default:"localhost:8080"`
	Port         int    `envconfig:"PORT" default:"8080"`
	Https        bool   `envconfig:"HTTPS" default:"false"`
}

func ParseConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("activitypub", &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}
	return &cfg, nil
}

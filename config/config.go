package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
)

type Config struct {
	PostgresConfig `envPrefix:"POSTGRES_"`
	ServiceConfig  `envPrefix:"SERVICE_"`
}

type ServiceConfig struct {
	TaskTimeout time.Duration `env:"TASK_TIMEOUT" envDefault:"1m"`
}

type PostgresConfig struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	DbName   string `env:"DB_NAME,required"`
}

func (pc PostgresConfig) String() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", pc.Host, pc.User, pc.Password, pc.DbName, pc.Port)
}

func BuildConfig() (*Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build cfg from env")
	}

	return &cfg, nil
}

package config

import "time"

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

package services

import (
	"time"
)

type EnvironmentServiceConfig struct {
	Path string `env:"APP_ENV"`
}

type SQLServiceConfig struct {
	DSN         string        `env:"DATABASE_URL"`
	MaxIdleConn int           `env:"DB_MAX_IDLE_CONN" envDefault:"10"`
	MaxOpenConn int           `env:"DB_MAX_OPEN_CONN" envDefault:"100"`
	MaxLifetime time.Duration `env:"DB_MAX_LIFETIME" envDefault:"1h"`
}

type StorageServiceConfig struct {
	AccessKey   string `env:"STORAGE_ACCESS_KEY"`
	SecretKey   string `env:"STORAGE_SECRET_KEY"`
	Bucket      string `env:"STORAGE_BUCKET"`
	Prefix      string `env:"STORAGE_PREFIX"`
	MaxFilezize int64  `env:"STORAGE_MAX_SIZE" envDefault:"10485760"`
}

type CacheServiceConfig struct {
	Host     string `env:"REDIS_HOST"`
	Password string `env:"REDIS_PASSWORD"`
	Username string `env:"REDIS_USERNAME"`
	Port     int    `env:"REDIS_PORT"`
}

type BrokerServiceConfig struct {
	Host string `env:"NATS_HOST"`
	Port int    `env:"NATS_CLIENT_PORT"`
}

type SecurityServiceConfig struct {
	Memory      uint32 `env:"PASSWORD_MEMORY"`
	Iterations  uint32 `env:"PASSWORD_ITERATIONS"`
	Parallelism uint8  `env:"PASSWORD_PARALLELISM"`
	SaltLength  uint32 `env:"PASSWORD_SALT_LENTH"`
	KeyLength   uint32 `env:"PASSWORD_KEY_LENGTH"`
	Secret      []byte `env:"PASSWORD_SECRET"`
}

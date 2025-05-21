package services

import (
	"time"
)

type SQLServiceConfig struct {
	DSN         string
	MaxIdleConn int
	MaxOpenConn int
	MaxLifetime time.Duration
}

type StorageServiceConfig struct {
	AccessKey   string
	SecretKey   string
	Bucket      string
	Prefix      string
	MaxFilezize int64
}

type CacheServiceConfig struct {
	Host     string
	Password string
	Username string
	Port     int
}

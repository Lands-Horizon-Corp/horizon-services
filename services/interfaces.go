package services

import "time"

type SQLServiceConfig struct {
	DSN         string
	MaxIdleConn int
	MaxOpenConn int
	MaxLifetime time.Duration
}

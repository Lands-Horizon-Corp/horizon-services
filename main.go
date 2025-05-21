package main

import (
	"context"
	"fmt"

	"github.com/lands-horizon/horizon-server/services"
)

func main() {
	horizon := services.NewHorizonService(services.HorizonServiceConfig{
		SQLConfig: &services.SQLServiceConfig{
			DSN:         "",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
			MaxLifetime: 60,
		},
		StorageConfig: &services.StorageServiceConfig{
			AccessKey:   "",
			SecretKey:   "",
			Prefix:      "",
			Bucket:      "",
			MaxFilezize: 1024 * 1024 * 10, // 10 MB
		},
		CacheConfig: &services.CacheServiceConfig{
			Host:     "",
			Password: "",
			Username: "",
			Port:     17356,
		},
		BrokerConfig: &services.BrokerServiceConfig{
			Host: "",
			Port: 4222,
		},
	})

	if err := horizon.Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
	}
}

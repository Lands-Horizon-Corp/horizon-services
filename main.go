package main

import (
	"context"
	"fmt"

	"github.com/lands-horizon/horizon-server/services"
)

func main() {
	horizon := services.NewHorizonService(services.HorizonServiceConfig{
		EnvironmentConfig: &services.EnvironmentServiceConfig{
			Path: "./.env",
		},
		SecurityConfig: &services.SecurityServiceConfig{
			Memory:      65536, // 64MB
			Iterations:  3,
			Parallelism: 2,  // 2 threads
			SaltLength:  16, // 16 bytes
			KeyLength:   32, // 32 bytes
			Secret:      []byte("your-secret-key"),
		},
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
		OTPServiceConfig: &services.OTPServiceConfig{
			Secret: []byte("your-otp-secret"),
		},
	})

	if err := horizon.Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
	}
}

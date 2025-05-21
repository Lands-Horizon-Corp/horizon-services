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
	})

	if err := horizon.Run(context.Background(), 1000); err != nil {
		fmt.Println("Error:", err)
	}
}

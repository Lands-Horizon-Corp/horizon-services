package main

import (
	"context"
	"fmt"

	"github.com/lands-horizon/horizon-server/services"
)

func main() {
	horizon := services.NewHorizonService()

	horizon.AddSQLDatabase("default", services.SQLServiceConfig{
		DSN:         "",
		MaxIdleConn: 10,
		MaxOpenConn: 100,
		MaxLifetime: 60,
	})

	if err := horizon.Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
	}
}

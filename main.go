package main

import (
	"context"
	"fmt"

	"github.com/lands-horizon/horizon-server/services"
)

func main() {

	horizon := services.NewHorizonService()

	if err := horizon.Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
	}
}

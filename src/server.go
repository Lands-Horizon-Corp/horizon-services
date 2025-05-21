package src

import (
	"context"

	"github.com/lands-horizon/horizon-server/services"
	"go.uber.org/fx"
)

type Provider struct {
	Service *services.HorizonService
}

func NewProvider(lc fx.Lifecycle) *Provider {
	horizonService := services.NewHorizonService(services.HorizonServiceConfig{
		EnvironmentConfig: &services.EnvironmentServiceConfig{
			Path: "./.env",
		},
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := horizonService.Run(ctx); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := horizonService.Stop(ctx); err != nil {
				return err
			}
			return nil
		},
	})
	return &Provider{
		Service: horizonService,
	}
}

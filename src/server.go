package src

import (
	"github.com/lands-horizon/horizon-server/services"
)

type Provider struct {
	Service *services.HorizonService
}

func NewProvider() *Provider {
	horizonService := services.NewHorizonService(services.HorizonServiceConfig{
		EnvironmentConfig: &services.EnvironmentServiceConfig{
			Path: "./.env",
		},
	})
	return &Provider{
		Service: horizonService,
	}
}

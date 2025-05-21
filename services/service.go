package services

import (
	"context"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

type HorizonService struct {
	Database map[string]horizon.SQLDatabase
}

func NewHorizonService() *HorizonService {
	return &HorizonService{
		Database: make(map[string]horizon.SQLDatabase),
	}
}

func (h *HorizonService) AddSQLDatabase(key string, config SQLServiceConfig) {
	database := horizon.NewGormDatabase(
		config.DSN,
		config.MaxIdleConn,
		config.MaxOpenConn,
		config.MaxLifetime,
	)
	h.Database[key] = database
}

func (h *HorizonService) Run(ctx context.Context) error {

	for _, db := range h.Database {
		if err := db.Run(ctx); err != nil {
			return err
		}
		if err := db.Ping(ctx); err != nil {
			return err
		}
	}
	return nil
}

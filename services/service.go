package services

import (
	"context"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

type HorizonService struct {
	Database horizon.SQLDatabaseService
	Storage  horizon.StorageService
	Cache    horizon.CacheService
}

func NewHorizonService(
	sqlConfig SQLServiceConfig,
	storageConfig StorageServiceConfig,
	cacheConfig CacheServiceConfig,
) *HorizonService {
	return &HorizonService{
		Database: horizon.NewGormDatabase(
			sqlConfig.DSN,
			sqlConfig.MaxIdleConn,
			sqlConfig.MaxOpenConn,
			sqlConfig.MaxLifetime,
		),
		Storage: horizon.NewHorizonStorageService(
			storageConfig.AccessKey,
			storageConfig.SecretKey,
			storageConfig.Prefix,
			storageConfig.Bucket,
			storageConfig.MaxFilezize,
		),
		Cache: horizon.NewHorizonCache(
			cacheConfig.Host,
			cacheConfig.Password,
			cacheConfig.Username,
			cacheConfig.Port,
		),
	}
}

func (h *HorizonService) Run(ctx context.Context) error {
	if err := h.Cache.Run(ctx); err != nil {
		return err
	}
	if err := h.Cache.Ping(ctx); err != nil {
		return err
	}
	if err := h.Storage.Run(ctx); err != nil {
		return err
	}
	if err := h.Database.Run(ctx); err != nil {
		return err
	}
	if err := h.Database.Ping(ctx); err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

type HorizonService struct {
	Database horizon.SQLDatabaseService
	Storage  horizon.StorageService
	Cache    horizon.CacheService
	Broker   horizon.MessageBrokerService
	Cron     horizon.SchedulerService
}

type HorizonServiceConfig struct {
	SQLConfig     *SQLServiceConfig
	StorageConfig *StorageServiceConfig
	CacheConfig   *CacheServiceConfig
	BrokerConfig  *BrokerServiceConfig
}

func NewHorizonService(cfg HorizonServiceConfig) *HorizonService {
	service := &HorizonService{}

	if cfg.SQLConfig != nil {
		service.Database = horizon.NewGormDatabase(
			cfg.SQLConfig.DSN,
			cfg.SQLConfig.MaxIdleConn,
			cfg.SQLConfig.MaxOpenConn,
			cfg.SQLConfig.MaxLifetime,
		)
	}

	if cfg.StorageConfig != nil {
		service.Storage = horizon.NewHorizonStorageService(
			cfg.StorageConfig.AccessKey,
			cfg.StorageConfig.SecretKey,
			cfg.StorageConfig.Prefix,
			cfg.StorageConfig.Bucket,
			cfg.StorageConfig.MaxFilezize,
		)
	}

	if cfg.CacheConfig != nil {
		service.Cache = horizon.NewHorizonCache(
			cfg.CacheConfig.Host,
			cfg.CacheConfig.Password,
			cfg.CacheConfig.Username,
			cfg.CacheConfig.Port,
		)
	}

	if cfg.BrokerConfig != nil {
		service.Broker = horizon.NewHorizonMessageBroker(
			cfg.BrokerConfig.Host,
			cfg.BrokerConfig.Port,
		)
	}

	// Scheduler (Cron) is always initialized if needed
	service.Cron = horizon.NewHorizonSchedule()

	return service
}

func (h *HorizonService) Run(ctx context.Context) error {
	if h.Cron != nil {
		if err := h.Cron.Run(ctx); err != nil {
			return err
		}
	}

	if h.Broker != nil {
		if err := h.Broker.Run(ctx); err != nil {
			return err
		}
	}

	if h.Cache != nil {
		if err := h.Cache.Run(ctx); err != nil {
			return err
		}
		if err := h.Cache.Ping(ctx); err != nil {
			return err
		}
	}

	if h.Storage != nil {
		if err := h.Storage.Run(ctx); err != nil {
			return err
		}
	}

	if h.Database != nil {
		if err := h.Database.Run(ctx); err != nil {
			return err
		}
		if err := h.Database.Ping(ctx); err != nil {
			return err
		}
	}

	return nil
}

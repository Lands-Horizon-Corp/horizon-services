package services

import (
	"context"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/rotisserie/eris"
)

type HorizonService struct {
	Environment horizon.EnvironmentService
	Database    horizon.SQLDatabaseService
	Storage     horizon.StorageService
	Cache       horizon.CacheService
	Broker      horizon.MessageBrokerService
	Cron        horizon.SchedulerService
	Security    horizon.SecurityService
	OTP         horizon.OTPService
}

type HorizonServiceConfig struct {
	EnvironmentConfig *EnvironmentServiceConfig
	SQLConfig         *SQLServiceConfig
	StorageConfig     *StorageServiceConfig
	CacheConfig       *CacheServiceConfig
	BrokerConfig      *BrokerServiceConfig
	SecurityConfig    *SecurityServiceConfig
	OTPServiceConfig  *OTPServiceConfig
}

func NewHorizonService(cfg HorizonServiceConfig) *HorizonService {
	service := &HorizonService{}
	if cfg.SecurityConfig != nil {
		service.Security = horizon.NewSecurityService(
			cfg.SecurityConfig.Memory,
			cfg.SecurityConfig.Iterations,
			cfg.SecurityConfig.Parallelism,
			cfg.SecurityConfig.SaltLength,
			cfg.SecurityConfig.KeyLength,
			cfg.SecurityConfig.Secret,
		)
	}
	if cfg.EnvironmentConfig != nil {
		service.Environment = horizon.NewEnvironmentService(
			cfg.EnvironmentConfig.Path,
		)
	}
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
	if cfg.OTPServiceConfig != nil {
		service.OTP = horizon.NewHorizonOTP(
			cfg.OTPServiceConfig.Secret,
			service.Cache,
			service.Security,
		)
	}
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
	if h.OTP != nil {
		if h.Cache == nil {
			return eris.New("OTP service requires a cache service")
		}
		if h.Security == nil {
			return eris.New("OTP service requires a security service")
		}
	}
	return nil
}

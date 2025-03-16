package db

import (
	"dice-game/pkg/config"
	"dice-game/pkg/domain/interfaces"
	"dice-game/pkg/domain/repository"
	"dice-game/pkg/infrastructure/db/postgresql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type StorageType string

const (
	PostgreSQL StorageType = "postgres"
)

func NewStore(storageType StorageType, cfg *config.AppConfig, logger interfaces.Logger) repository.DataStore {
	switch storageType {
	case PostgreSQL:
		return postgresql.NewPostgresStore(cfg, logger)
	default:
		logger.Warn().Str("type", string(storageType)).Msg("Unknown storage type, using PostgreSQL")
		return postgresql.NewPostgresStore(cfg, logger)
	}
}

package db

import (
	"dice-game/pkg/config"
	"dice-game/pkg/domain/repository"
	"dice-game/pkg/infrastructure/db/postgresql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

func NewStore(cfg *config.AppConfig, logger *zerolog.Logger) repository.DataStore {
	return postgresql.NewPostgresStore(cfg, logger)
}

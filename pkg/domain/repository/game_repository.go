package repository

import (
	"context"
	"dice-game/pkg/domain/model"
)

type DataStore interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	RunMigrations(migrationsPath string) error
	WithTransaction(ctx context.Context, txFunc func(tx Transaction) error) error
	GetGameRepository() GameRepository
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type GameRepository interface {
	SaveGameResult(ctx context.Context, result *model.GameResult) error
	GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error)
	GetGameResultsByPlayer(ctx context.Context, playerID string, limit, offset int) ([]*model.GameResult, error)
	GetTotalGames(ctx context.Context) (int, error)
}

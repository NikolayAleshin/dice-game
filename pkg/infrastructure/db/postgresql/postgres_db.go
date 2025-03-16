package postgresql

import (
	"context"
	"database/sql"
	"dice-game/pkg/config"
	"dice-game/pkg/domain/model"
	"dice-game/pkg/domain/repository"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"sync"
	"time"
)

var _ repository.DataStore = (*PostgresStore)(nil)

type PostgresStore struct {
	sync.RWMutex
	pool   *pgxpool.Pool
	config *config.AppConfig
	logger zerolog.Logger

	gameRepo *PostgresGameRepository
}

func NewPostgresStore(cfg *config.AppConfig, logger *zerolog.Logger) *PostgresStore {
	return &PostgresStore{
		config: cfg,
		logger: logger.With().Str("component", "postgresql-database").Logger(),
	}
}

func (s *PostgresStore) Connect(ctx context.Context) error {
	s.logger.Info().Msg("Connecting to PostgreSQL database...")

	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	connString := s.config.DatabaseURL()
	if connString == "" {
		connString = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			s.config.Database.User,
			s.config.Database.Password,
			s.config.Database.Host,
			s.config.Database.Port,
			s.config.Database.Name,
			s.config.Database.SSLMode,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return errors.Wrap(err, "failed to parse database connection string")
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(connCtx, poolConfig)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	if err := pool.Ping(connCtx); err != nil {
		pool.Close()
		return errors.Wrap(err, "failed to ping database")
	}

	s.pool = pool

	s.gameRepo = &PostgresGameRepository{
		pool:   s.pool,
		logger: s.logger.With().Str("repository", "game").Logger(),
	}

	s.logger.Info().Msg("Successfully connected to PostgreSQL database")
	return nil
}

func (s *PostgresStore) Close(ctx context.Context) error {
	if s.pool != nil {
		s.logger.Info().Msg("Closing database connection...")
		s.pool.Close()
		s.logger.Info().Msg("Database connection closed")
	}
	return nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	if s.pool == nil {
		return errors.New("database connection is not initialized")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.pool.Ping(pingCtx)
}

func (s *PostgresStore) RunMigrations(migrationsPath string) error {
	s.logger.Info().Str("path", migrationsPath).Msg("Running database migrations...")

	if s.pool == nil {
		return errors.New("database connection is not initialized")
	}

	connString := s.config.DatabaseURL()
	if connString == "" {
		connString = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			s.config.Database.User,
			s.config.Database.Password,
			s.config.Database.Host,
			s.config.Database.Port,
			s.config.Database.Name,
			s.config.Database.SSLMode,
		)
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return errors.Wrap(err, "failed to open database connection for migrations")
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to close database connection")
		}
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create migration driver")
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create migrator")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "failed to run migrations")
	}

	s.logger.Info().Msg("Database migrations completed successfully")
	return nil
}

type PostgresTransaction struct {
	tx pgx.Tx
}

func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (s *PostgresStore) WithTransaction(ctx context.Context, txFunc func(tx repository.Transaction) error) error {
	if s.pool == nil {
		return errors.New("database connection is not initialized")
	}

	pgxTx, err := s.pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	tx := &PostgresTransaction{tx: pgxTx}

	if err := txFunc(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Wrap(rbErr, fmt.Sprintf("failed to rollback transaction after error: %v", err))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (s *PostgresStore) GetGameRepository() repository.GameRepository {
	return s.gameRepo
}

type PostgresGameRepository struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

var _ repository.GameRepository = (*PostgresGameRepository)(nil)

func (r *PostgresGameRepository) SaveGameResult(ctx context.Context, result *model.GameResult) error {
	if r.pool == nil {
		return errors.New("database connection is not initialized")
	}

	query := `
		INSERT INTO game_results (
			game_id, player_id, player_dice, server_dice, 
			winner, played_at, generator_used, verification_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		result.GameID,
		result.PlayerID,
		result.PlayerDice,
		result.ServerDice,
		string(result.Winner),
		result.PlayedAt,
		result.GeneratorUsed,
		result.VerificationKey,
	)

	if err != nil {
		return errors.Wrap(err, "failed to save game result")
	}

	return nil
}

func (r *PostgresGameRepository) GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error) {
	if r.pool == nil {
		return nil, errors.New("database connection is not initialized")
	}

	query := `
		SELECT 
			game_id, player_id, player_dice, server_dice, 
			winner, played_at, generator_used, verification_key
		FROM game_results
		WHERE game_id = $1
	`

	var result model.GameResult
	var winner string
	var playedAt time.Time

	err := r.pool.QueryRow(ctx, query, gameID).Scan(
		&result.GameID,
		&result.PlayerID,
		&result.PlayerDice,
		&result.ServerDice,
		&winner,
		&playedAt,
		&result.GeneratorUsed,
		&result.VerificationKey,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("game not found")
		}
		return nil, errors.Wrap(err, "failed to get game result")
	}

	result.Winner = model.Winner(winner)
	result.PlayedAt = playedAt

	return &result, nil
}

func (r *PostgresGameRepository) GetGameResultsByPlayer(ctx context.Context, playerID string, limit, offset int) ([]*model.GameResult, error) {
	if r.pool == nil {
		return nil, errors.New("database connection is not initialized")
	}

	query := `
		SELECT 
			game_id, player_id, player_dice, server_dice, 
			winner, played_at, generator_used, verification_key
		FROM game_results
		WHERE player_id = $1
		ORDER BY played_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, playerID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query game results")
	}
	defer rows.Close()

	var results []*model.GameResult

	for rows.Next() {
		var result model.GameResult
		var winner string
		var playedAt time.Time

		err := rows.Scan(
			&result.GameID,
			&result.PlayerID,
			&result.PlayerDice,
			&result.ServerDice,
			&winner,
			&playedAt,
			&result.GeneratorUsed,
			&result.VerificationKey,
		)

		if err != nil {
			return nil, errors.Wrap(err, "failed to scan game result")
		}

		result.Winner = model.Winner(winner)
		result.PlayedAt = playedAt
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating game results")
	}

	return results, nil
}

func (r *PostgresGameRepository) GetTotalGames(ctx context.Context) (int, error) {
	if r.pool == nil {
		return 0, errors.New("database connection is not initialized")
	}

	query := `SELECT COUNT(*) FROM game_results`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get total games count")
	}

	return count, nil
}

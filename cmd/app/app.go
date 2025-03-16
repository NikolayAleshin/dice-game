package app

import (
	"context"
	"dice-game/pkg/config"
	"dice-game/pkg/domain/repository"
	"dice-game/pkg/domain/service"
	"dice-game/pkg/infrastructure/db"
	"dice-game/pkg/infrastructure/grpc"
	"dice-game/pkg/infrastructure/random"
	"dice-game/pkg/usecase"
	"fmt"
	"github.com/rs/zerolog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Application struct {
	once          sync.Once
	logger        *zerolog.Logger
	config        *config.AppConfig
	initialized   bool
	configMutex   sync.Mutex
	dataStore     repository.DataStore
	grpcServer    *grpc.Server
	randomService service.RandomServiceInterface
	gameService   service.GameServiceInterface
	gameUseCase   usecase.GameUseCaseInterface
}

func NewApplication() *Application {
	initialLogger := createDefaultLogger()

	app := &Application{
		logger: initialLogger,
	}

	return app
}

func NewZerologAdapter(level string, isJSON bool) *zerolog.Logger {
	var zerologLevel zerolog.Level

	switch level {
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "info":
		zerologLevel = zerolog.InfoLevel
	case "warn":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	var logger zerolog.Logger

	if isJSON {
		logger = zerolog.New(os.Stdout).
			Level(zerologLevel).
			With().
			Timestamp().
			Logger()
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		}

		logger = zerolog.New(output).
			Level(zerologLevel).
			With().
			Timestamp().
			Logger()
	}

	return &logger
}

func createDefaultLogger() *zerolog.Logger {
	return NewZerologAdapter("info", false)
}

func (a *Application) Run() {
	a.once.Do(func() {
		a.logger.Info().Msg("Application is starting...")

		if err := a.LoadConfig(); err != nil {
			a.logger.Fatal().Err(err).Msg("Failed to load configuration")
		}

		a.configureLogger()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		a.logger.Info().Msg("Application is running...")

		g, gCtx := errgroup.WithContext(ctx)

		g.Go(func() error {
			return a.Start(gCtx)
		})

		if err := g.Wait(); err != nil {
			a.logger.Error().Err(err).Msg("Application stopped with error")
		} else {
			a.logger.Info().Msg("Application stopped gracefully")
		}

		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()

		if err := a.Stop(stopCtx); err != nil {
			a.logger.Error().Err(err).Msg("Error during application shutdown")
		}
	})
}

func (a *Application) LoadConfig() error {
	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	if a.initialized {
		return nil
	}

	a.logger.Debug().Msg("Loading configuration...")

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("$HOME/.dice-game")
	v.AddConfigPath(".")

	v.SetEnvPrefix("DICE")
	v.AutomaticEnv()

	dbEnvMappings := map[string]string{
		"database.url":             "DATABASE_URL",
		"database.user":            "DATABASE_USER",
		"database.password":        "DATABASE_PASSWORD",
		"database.type":            "DATABASE_TYPE",
		"database.host":            "DATABASE_HOST",
		"database.port":            "DATABASE_PORT",
		"database.name":            "DATABASE_NAME",
		"database.ssl_mode":        "DATABASE_SSL_MODE",
		"database.run_migrations":  "DATABASE_RUN_MIGRATIONS",
		"database.migrations_path": "DATABASE_MIGRATIONS_PATH",
	}

	for configKey, envVar := range dbEnvMappings {
		v.BindEnv(configKey, envVar)
	}

	v.BindEnv("grpc.host", "GRPC_HOST")
	v.BindEnv("http.port", "HTTP_PORT")
	v.BindEnv("http.host", "HTTP_HOST")
	v.BindEnv("log.level", "LOG_LEVEL")
	v.BindEnv("log.json", "LOG_JSON")
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("version", "VERSION")
	v.BindEnv("game.enable_verification", "GAME_ENABLE_VERIFICATION")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return errors.Wrap(err, "failed to read config file")
		}
		a.logger.Warn().Msg("Config file not found, using environment variables and defaults")
	} else {
		a.logger.Info().Str("file", v.ConfigFileUsed()).Msg("Config file loaded")
	}

	if err := v.Unmarshal(&a.config); err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}

	if err := a.validateConfig(); err != nil {
		return errors.Wrap(err, "config validation failed")
	}

	a.initialized = true
	a.logger.Debug().Msg("Configuration loaded successfully")
	return nil
}

func (a *Application) validateConfig() error {
	a.logger.Debug().Interface("database", a.config.Database).Msg("Validating database configuration")
	a.logger.Debug().
		Str("host", a.config.Database.Host).
		Str("port", a.config.Database.Port).
		Str("user", a.config.Database.User).
		Str("database", a.config.Database.Name).
		Msg("Database connection details")

	if a.config.Database.URL == "" {
		if a.config.Database.Host == "" {
			a.logger.Error().Msg("Database host is empty")
			return fmt.Errorf("database host is not configured")
		}
		if a.config.Database.Port == "" {
			a.logger.Error().Msg("Database port is empty")
			return fmt.Errorf("database port is not configured")
		}
		if a.config.Database.Name == "" {
			a.logger.Error().Msg("Database name is empty")
			return fmt.Errorf("database name is not configured")
		}
		if a.config.Database.User == "" {
			a.logger.Error().Msg("Database user is empty")
			return fmt.Errorf("database user is not configured")
		}
		if a.config.Database.Password == "" {
			a.logger.Error().Msg("Database password is empty")
			return fmt.Errorf("database password is not configured")
		}
	}

	a.logger.Debug().Interface("grpc", a.config.GRPC).Msg("Validating GRPC configuration")

	if a.config.GRPC.Port == "" {
		a.logger.Error().Msg("GRPC port is empty")
		return fmt.Errorf("GRPC port is not configured")
	}

	return nil
}

func (a *Application) configureLogger() {
	a.logger = NewZerologAdapter(a.config.Log.Level, a.config.Log.JSON)
}

func (a *Application) GetConfig() *config.AppConfig {
	return a.config
}

func (a *Application) Start(ctx context.Context) error {
	a.logger.Info().Msg("Starting application components...")

	if err := a.initDatabase(ctx); err != nil {
		return err
	}

	a.initRandomGenerators()
	a.initServices()

	if err := a.startGRPCServer(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

func (a *Application) initDatabase(ctx context.Context) error {

	a.dataStore = db.NewStore(a.config, a.logger)
	if err := a.dataStore.Connect(ctx); err != nil {
		a.logger.Error().Err(err).Msg("Failed to connect to database")
		return errors.Wrap(err, "failed to connect to database")
	}

	if a.config.Database.RunMigrations {
		migrationsPath := a.config.Database.MigrationsPath
		if migrationsPath == "" {
			migrationsPath = "migrations"
			a.logger.Warn().Msg("Migrations path not set, using default: 'migrations'")
		}

		a.logger.Info().
			Str("path", migrationsPath).
			Msg("Running database migrations...")

		if err := a.dataStore.RunMigrations(migrationsPath); err != nil {
			a.logger.Error().
				Err(err).
				Str("path", migrationsPath).
				Msg("Failed to run database migrations")
			return errors.Wrap(err, "failed to run database migrations")
		}

		a.logger.Info().Msg("Database migrations completed successfully")
	} else {
		a.logger.Info().Msg("Database migrations skipped (disabled in config)")
	}

	return nil
}

func (a *Application) initRandomGenerators() {
	randomGenerators := []random.Generator{
		random.NewStandardGenerator(),
		random.NewCryptoGenerator(),
	}

	if a.config.Game.EnableVerification {
		clientSeedFunc := func() string {
			return time.Now().Format(time.RFC3339)
		}

		serverSeed := fmt.Sprintf("server-seed-%d", time.Now().UnixNano())

		provablyFairGen := random.NewProovablyFairGenerator(serverSeed, clientSeedFunc)
		randomGenerators = append(randomGenerators, provablyFairGen)
	}

	a.randomService = service.NewRandomService(randomGenerators)
}

func (a *Application) initServices() {
	gameRepository := a.dataStore.GetGameRepository()
	a.gameService = service.NewGameService(a.randomService, gameRepository)
	a.gameUseCase = usecase.NewGameUseCase(a.gameService)
}

func (a *Application) startGRPCServer(ctx context.Context) error {
	grpcAddr := net.JoinHostPort(a.config.GRPCHost(), a.config.GRPCPort())
	a.grpcServer = grpc.NewServer(grpcAddr, a.logger, a.gameUseCase)
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		a.logger.Info().Str("address", grpcAddr).Msg("Starting gRPC server")
		err := a.grpcServer.Start()
		if err != nil {
			a.logger.Error().Err(err).Msg("gRPC server failed to start")
		}
		return err
	})

	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-gCtx.Done():
		return g.Wait()
	}
}

func (a *Application) Stop(ctx context.Context) error {
	a.logger.Info().Msg("Shutting down application components...")

	if a.grpcServer != nil {
		a.logger.Info().Msg("Stopping gRPC server...")
		a.grpcServer.Stop()
	}

	if a.dataStore != nil {
		a.logger.Info().Msg("Closing database connection...")
		if err := a.dataStore.Close(ctx); err != nil {
			a.logger.Error().Err(err).Msg("Error closing database connection")
			return errors.Wrap(err, "failed to close database connection")
		}
	}

	return nil
}

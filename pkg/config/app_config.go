package config

type AppConfig struct {
	HTTP        HTTPConfig     `mapstructure:"http"`
	Database    DatabaseConfig `mapstructure:"database"`
	Log         LogConfig      `mapstructure:"log"`
	Environment string         `mapstructure:"environment"`
	Version     string         `mapstructure:"version"`
	GRPC        GRPCConfig     `mapstructure:"grpc"`
	Game        GameConfig     `mapstructure:"game"`
}

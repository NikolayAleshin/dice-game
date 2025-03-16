package config

type GameConfig struct {
	DefaultGeneratorType string `mapstructure:"default_generator_type"`
	EnableVerification   bool   `mapstructure:"enable_verification"`
}

package config

import (
	"fmt"
)

type DatabaseConfig struct {
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	URL            string `mapstructure:"url"`
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	Name           string `mapstructure:"name"`
	SSLMode        string `mapstructure:"ssl_mode"`
	RunMigrations  bool   `mapstructure:"run_migrations"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

func (c *AppConfig) DatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	if c.Database.Host != "" && c.Database.Name != "" {
		return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			c.Database.User, c.Database.Password, c.Database.Host,
			c.Database.Port, c.Database.Name, c.Database.SSLMode)
	}
	return ""
}

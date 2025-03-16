package config

type HTTPConfig struct {
	Port string
	Host string
}

func (c *AppConfig) BaseURI(res string) string {
	return c.HTTPHost() + res
}

func (c *AppConfig) HTTPHost() string {
	if c.HTTP.Host == "" {
		return "0.0.0.0"
	}
	return c.HTTP.Host
}

func (c *AppConfig) HTTPPort() string {
	if c.HTTP.Port == "" {
		return "2342"
	}
	return c.HTTP.Port
}

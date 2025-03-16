package config

type GRPCConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

func (c *AppConfig) GRPCHost() string {
	if c.GRPC.Host == "" {
		return "0.0.0.0"
	}
	return c.GRPC.Host
}

func (c *AppConfig) GRPCPort() string {
	if c.GRPC.Port == "" {
		return "9090"
	}
	return c.GRPC.Port
}

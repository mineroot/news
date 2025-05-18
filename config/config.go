package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
)

type config struct {
	Env            string `env:"APP_ENV" validate:"required,oneof=dev prod"`
	LogLevel       string `env:"APP_LOG_LEVEL" validate:"required,oneof=debug info warning error"`
	MongoURI       string `env:"APP_MONGO_URI" validate:"required"`
	HttpServerPort string `env:"APP_HTTP_SERVER_PORT" validate:"required,alphanum"`
}

type Config struct {
	config config
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v", c.config)
}

func (c *Config) IsProd() bool {
	return c.config.Env == "prod"
}

func (c *Config) MongoUri() string {
	return c.config.MongoURI
}

func (c *Config) HttpServerPort() string {
	return c.config.HttpServerPort
}

func (c *Config) LogLevel() zerolog.Level {
	level, err := zerolog.ParseLevel(c.config.LogLevel)
	if err != nil {
		return zerolog.InfoLevel
	}
	return level
}

func LoadConfig() (*Config, error) {
	cfg := config{
		LogLevel: zerolog.LevelInfoValue,
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config: unable to read env variables: %w", err)
	}
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(cfg); err != nil {
		return nil, fmt.Errorf("config: config has invalid values: %w", err)
	}

	return &Config{config: cfg}, nil
}

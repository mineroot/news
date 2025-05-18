package config_test

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/mineroot/news/config"
)

func setEnv(vars map[string]string) func() {
	originals := make(map[string]string)
	for k, v := range vars {
		originals[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}
	return func() {
		for k, v := range originals {
			_ = os.Setenv(k, v)
		}
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	restore := setEnv(map[string]string{
		"APP_ENV":              "prod",
		"APP_LOG_LEVEL":        "info",
		"APP_MONGO_URI":        "mongodb://localhost:27017",
		"APP_HTTP_SERVER_PORT": "8080",
	})
	defer restore()

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.True(t, cfg.IsProd())
	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoUri())
	assert.Equal(t, "8080", cfg.HttpServerPort())
	assert.Equal(t, zerolog.InfoLevel, cfg.LogLevel())
}

func TestLoadConfig_InvalidEnv(t *testing.T) {
	restore := setEnv(map[string]string{
		"APP_ENV":              "invalid",
		"APP_LOG_LEVEL":        "info",
		"APP_MONGO_URI":        "uri",
		"APP_HTTP_SERVER_PORT": "8080",
	})
	defer restore()

	_, err := config.LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfig_MissingValues(t *testing.T) {
	restore := setEnv(map[string]string{})
	defer restore()

	_, err := config.LoadConfig()
	assert.Error(t, err)
}

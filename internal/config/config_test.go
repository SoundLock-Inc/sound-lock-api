package config_test

import (
	"os"
	"path/filepath"
	"sound_lock/internal/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	return path
}

func TestMustLoadByPath_ValidConfig(t *testing.T) {
	yaml := `env: "dev"
access_token_ttl: 15m
refresh_token_ttl: 168h
accessTokenSigningKey: "accessTokenSigningKey"
refreshTokenSigningKey: "refreshTokenSigningKey"
port: 8080
host: "localhost"

db:
  db_host: "localhost"
  db_port: "5432"
  db_password: "secret"
  db_name: "sound_lock"
  ssl_mode: "disable"
  username: "admin"
`
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DBNAME", "testdb")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("DBNAME")

	path := createTempConfig(t, yaml)

	cfg := config.MustLoadByPath(path)

	assert.Equal(t, "dev", cfg.Env)
	assert.Equal(t, "5432", cfg.DB.DBPort)
	assert.Equal(t, "secret", cfg.DB.Password)
	assert.Equal(t, "testdb", cfg.DB.DBName)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 15*time.Minute, cfg.AccessTokenTTL)
}

func TestMustLoadByPath_ConfigFileNotFound(t *testing.T) {
	fakePath := "fake_path/config.yaml"

	defer func() {
		r := recover()
		require.NotNil(t, r)
		require.Contains(t, r, "config file not found")
	}()

	_ = config.MustLoadByPath(fakePath)
}

func TestMustLoadByPath_InvalidYAML(t *testing.T) {
	badYAML := `
env: "dev"
access_token_ttl: 15m
db:
  db_port: "5432"
  bad_indent
`

	path := createTempConfig(t, badYAML)

	assert.Panics(t, func() {
		config.MustLoadByPath(path)
	})
}

func TestFetchConfigPath_FromEnv(t *testing.T) {
	yaml := `
access_token_ttl: 10m
refresh_token_ttl: 24h
db:
  db_port: "5432"
  ssl_mode: "disable"
  username: "admin"
  db_name: "prod"
  db_host: "localhost"
`

	os.Setenv("CONFIG_PATH", createTempConfig(t, yaml))
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("DB_PASSWORD", "topsecret")
	defer os.Unsetenv("DB_PASSWORD")

	cfg := config.MustLoad()
	assert.Equal(t, "local", cfg.Env)
	assert.Equal(t, "topsecret", cfg.DB.Password)
}

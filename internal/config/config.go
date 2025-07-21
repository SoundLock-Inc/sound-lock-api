package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config содержит все конфигурационные параметры сервиса.
type Config struct {
	Env                    string        `yaml:"env" env-default:"local"`
	AccessTokenTTL         time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL        time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
	AccessTokenSigningKey  string        `yaml:"accessTokenSigningKey"`
	RefreshTokenSigningKey string        `yaml:"refreshTokenSigningKey"`
	Port                   int           `yaml:"port"`
	Host                   string        `yaml:"host"`
	DB                     DBConfig      `yaml:"db"`
}

// DBConfig определяет параметры подключения к базе данных.
type DBConfig struct {
	DBPort   string `yaml:"db_port"`                       // Порт БД
	SSLMode  string `yaml:"ssl_mode"`                      // Режим SSL-соединения
	Username string `yaml:"username"`                      // Имя пользователя БД
	Password string `env:"DB_PASSWORD" yaml:"db_password"` // Пароль БД (загружается из переменной окружения)
	DBName   string `env:"DBNAME" yaml:"db_name"`          // Имя базы данных (может быть переопределено через ENV)
	DBHost   string `yaml:"db_host"`                       // Адрес хоста БД
}

// MustLoad загружает конфигурацию из файла или завершает работу при ошибке.
// Функция ищет путь к конфигурационному файлу через флаги командной строки
// или переменные окружения. Если путь не указан - вызывает панику.
func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

// MustLoadByPath загружает конфигурацию из конкретного файла.
// Проверяет существование файла, парсит конфигурацию и возвращает объект Config.
// Вызывает панику при любой ошибке загрузки или парсинга.
func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found: " + path)
	}

	var config Config

	err := cleanenv.ReadConfig(path, &config)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &config

}

// fetchConfigPath извлекает путь к конфигурационному файлу из:
// 1. Флагов командной строки (--config_path)
// 2. Переменной окружения CONFIG_PATH
// Если ни один источник не указан - возвращает пустую строку.
func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config_path", "", "Path to config path")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	return path
}

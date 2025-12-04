package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sound_lock/internal/config"
	"sound_lock/internal/handlers"
	"sound_lock/internal/http"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/pkg"
	"sound_lock/internal/service"
	"sound_lock/internal/storage"
	"sound_lock/internal/storage/postgres"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/labstack/echo/v4"
)

type App struct {
	*http.Server
}

func New(config *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := pkg.SetUpLogger()

	log.Debug("sldsd")

	e := echo.New()

	srv := http.New(ctx, config.Port, config.Host, e)

	db := postgres.New(ctx, config)

	authStorage := storage.NewUser(db)
	refreshStorage := storage.NewRefreshToken(db)

	jwt := jwt.NewJWTManager(
		config.AccessTokenSigningKey,
		config.RefreshTokenSigningKey,
		config.AccessTokenTTL,
		config.RefreshTokenTTL,
	)

	authService := service.NewAuth(
		authStorage,
		refreshStorage,
		jwt,
		config.AccessTokenTTL,
		config.RefreshTokenTTL,
	)

	profileService := service.NewProfile(authStorage)

	h := handlers.New(e, authService, profileService, config)

	h.SetupHandlers(config.AccessTokenSigningKey)

	return &App{srv}
}

// GeneratePasswordFromWAV принимает путь к WAV-файлу и длину желаемого пароля (8, 16, 32)
func GeneratePasswordFromWAV(path string, length int) (string, error) {
	if length != 8 && length != 16 && length != 32 {
		return "", fmt.Errorf("unsupported password length: %d", length)
	}

	// Открытие .wav файла
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// WAV decoder
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return "", fmt.Errorf("invalid WAV file")
	}

	// Чтение всех сэмплов
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return "", fmt.Errorf("failed to read audio buffer: %w", err)
	}

	// Преобразование PCM в байты
	raw := pcmToBytes(buf)

	// Хеширование PCM-данных
	hash := sha256.Sum256(raw)

	// Преобразуем хеш в строку
	hashStr := hex.EncodeToString(hash[:])

	// Вернём нужное количество символов
	if length > len(hashStr) {
		return "", fmt.Errorf("hash output too short")
	}
	return hashStr[:length], nil
}

// Преобразует PCM-сэмплы в байтовый массив (Little Endian)
func pcmToBytes(buf *audio.IntBuffer) []byte {
	data := buf.Data
	bytes := make([]byte, len(data)*2) // 16 бит на сэмпл = 2 байта
	for i, sample := range data {
		bytes[i*2] = byte(sample)          // младший байт
		bytes[i*2+1] = byte((sample >> 8)) // старший байт
	}
	return bytes
}

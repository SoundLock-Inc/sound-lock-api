package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sound_lock/internal/app"
	"sound_lock/internal/config"
	"syscall"
	"time"
)

// main является точкой входа в приложение.
//  1. Загружает конфигурацию.
//  2. Инициализирует и запускает приложение.
//  3. Ожидает сигнала завершения (SIGINT или SIGTERM).
//  4. Корректно завершает работу приложения с таймаутом в 10 секунд.
func main() {
	// Загружаем конфигурацию приложения.
	cfg := config.MustLoad()

	// Создаем экземпляр приложения.
	app := app.New(cfg)

	// Запускаем приложение в отдельной горутине.
	go app.MustRun()

	// Канал для получения системных сигналов.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем выполнение до получения сигнала завершения.
	<-quit

	slog.Info("shutdown initiated...")

	// Создаем контекст с таймаутом для корректной остановки приложения.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем приложение.
	app.Stop(shutdownCtx)
	slog.Info("server stopped...")
}

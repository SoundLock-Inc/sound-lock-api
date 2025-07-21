package http

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
)

// Server — структура, описывающая HTTP-сервер.
// Включает контекст для graceful shutdown, порт, хост и экземпляр Echo.
type Server struct {
	ctx  context.Context // Контекст для управления жизненным циклом сервера
	port int             // Порт, на котором будет запущен сервер
	host string          // Хост (обычно "0.0.0.0" или "localhost")
	e    *echo.Echo      // Инстанс Echo — основа HTTP-сервера
}

// New — конструктор для создания нового сервера.
// Принимает контекст, порт, хост и экземпляр Echo, возвращает *Server.
func New(
	ctx context.Context,
	port int,
	host string,
	e *echo.Echo,
) *Server {
	return &Server{
		ctx:  ctx,
		port: port,
		host: host,
		e:    e,
	}
}

// MustRun — запускает HTTP-сервер и завершает приложение с фатальной ошибкой, если запуск невозможен.
// Формирует строку адреса (хост:порт) и запускает Echo.
func (s *Server) MustRun() {
	host := fmt.Sprintf("%s:%d", s.host, s.port)
	s.e.Logger.Fatal(s.e.Start(host))
}

// Stop — корректно завершает работу сервера (graceful shutdown) с использованием контекста.
// Вызывается при завершении приложения или получении сигнала завершения.
func (s *Server) Stop(ctx context.Context) {
	if err := s.e.Shutdown(ctx); err != nil {
		s.e.Logger.Error("Can't stop server")
	}
}

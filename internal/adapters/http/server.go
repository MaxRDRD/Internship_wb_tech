package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"topq/internal/usecase"
)

type Server struct {
	httpServer *http.Server
}

// Создание нового HTTP сервера
func NewServer(addr string, top *usecase.Top, stopList *usecase.StopList, defaultTopN int, windowSeconds int) *Server {
	mux := http.NewServeMux()                                        // Создаем ServeMux для регистрации маршрутов
	handler := NewHandler(top, stopList, defaultTopN, windowSeconds) // Создаем Handler, который будет обрабатывать HTTP запросы
	handler.Register(mux)                                            // Регистрируем маршруты и их обработчики в ServeMux

	// Задаем конфигурацию HTTP сервера
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{httpServer: srv}
}

// Запуск HTTP сервера
func (s *Server) Run(ctx context.Context) error {
	// Канал для получения ошибок от сервера
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()
	// Ожидаем либо сигнала завершения, либо ошибки от сервера
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

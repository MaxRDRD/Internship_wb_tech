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

func NewServer(addr string, top *usecase.Top, stopList *usecase.StopList, defaultTopN int, windowSeconds int) *Server {
	mux := http.NewServeMux()
	handler := NewHandler(top, stopList, defaultTopN, windowSeconds)
	handler.Register(mux)

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

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

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

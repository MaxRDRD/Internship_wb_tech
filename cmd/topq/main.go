package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"topq/internal/adapters/http"
	"topq/internal/adapters/kafka"
	"topq/internal/adapters/memory"
	"topq/internal/config"
	"topq/internal/usecase"
)

func main() {
	// Load .env file into environment (if present) so developers can run
	// `go run ./cmd/topq` without exporting env vars manually.
	_ = godotenv.Load()

	cfg := config.Load()

	repo := memory.NewSlidingWindowRepo(cfg.WindowSeconds)
	stopRepo := memory.NewStopListRepo()
	stopList := usecase.NewStopList(stopRepo)
	ingest := usecase.NewIngest(repo, stopRepo)
	top := usecase.NewTop(repo, stopRepo)

	server := http.NewServer(cfg.HTTPAddr, top, stopList, cfg.DefaultTopN, cfg.WindowSeconds)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.ConsumerEnabled {
		consumer := kafka.NewConsumer(cfg, ingest)
		go func() {
			if err := consumer.Run(ctx); err != nil {
				log.Printf("kafka consumer stopped: %v", err)
			}
		}()
	} else {
		log.Printf("kafka consumer disabled: set KAFKA_BROKERS and KAFKA_TOPIC")
	}

	if err := server.Run(ctx); err != nil {
		log.Printf("http server stopped: %v", err)
	}

	<-ctx.Done()
	time.Sleep(50 * time.Millisecond)
}
